// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.8

package xorm

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"

	"xorm.io/core"
)

const (
	// Version show the xorm's version
	Version string = "0.7.0.0504"
)

//默认支持的数据库驱动
func regDrvsNDialects() bool {
	providedDrvsNDialects := map[string]struct {
		dbType     core.DbType
		getDriver  func() core.Driver
		getDialect func() core.Dialect
	}{
		"mssql":    {"mssql", func() core.Driver { return &odbcDriver{} }, func() core.Dialect { return &mssql{} }},
		"odbc":     {"mssql", func() core.Driver { return &odbcDriver{} }, func() core.Dialect { return &mssql{} }}, // !nashtsai! TODO change this when supporting MS Access
		"mysql":    {"mysql", func() core.Driver { return &mysqlDriver{} }, func() core.Dialect { return &mysql{} }},
		"mymysql":  {"mysql", func() core.Driver { return &mymysqlDriver{} }, func() core.Dialect { return &mysql{} }},
		"postgres": {"postgres", func() core.Driver { return &pqDriver{} }, func() core.Dialect { return &postgres{} }},
		"pgx":      {"postgres", func() core.Driver { return &pqDriverPgx{} }, func() core.Dialect { return &postgres{} }},
		"sqlite3":  {"sqlite3", func() core.Driver { return &sqlite3Driver{} }, func() core.Dialect { return &sqlite3{} }},
		"oci8":     {"oracle", func() core.Driver { return &oci8Driver{} }, func() core.Dialect { return &oracle{} }},
		"goracle":  {"oracle", func() core.Driver { return &goracleDriver{} }, func() core.Dialect { return &oracle{} }},
	}

	for driverName, v := range providedDrvsNDialects {
		//如果没有注册过, 就注册(存放到map)
		if driver := core.QueryDriver(driverName); driver == nil {
			core.RegisterDriver(driverName, v.getDriver())
			core.RegisterDialect(v.dbType, v.getDialect)
		}
	}
	return true
}

func close(engine *Engine) {
	engine.Close()
}

func init() {
	//注册默认数据库驱动方言,其实就是go struct和各个数据库的数据类型的转换
	regDrvsNDialects()
}

// NewEngine new a db manager according to the parameter. Currently support four
// drivers
//创建xorm的控制对象, 用于创建session, 发起查询等
func NewEngine(driverName string, dataSourceName string) (*Engine, error) {
	//查询驱动是否被支持, 在init时注册
	driver := core.QueryDriver(driverName)
	if driver == nil {
		return nil, fmt.Errorf("Unsupported driver name: %v", driverName)
	}

	//解析获得连接驱动的uri
	uri, err := driver.Parse(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	//根据驱动类型来获取转换关系, go struct和数据库类型的转换, 和一些特定的实现
	dialect := core.QueryDialect(uri.DbType)
	if dialect == nil {
		return nil, fmt.Errorf("Unsupported dialect type: %v", uri.DbType)
	}

	//打开数据库
	db, err := core.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	//初始化转换关系对象
	err = dialect.Init(db, uri, driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	engine := &Engine{
		db:             db,
		dialect:        dialect,
		Tables:         make(map[reflect.Type]*core.Table),
		mutex:          &sync.RWMutex{},
		TagIdentifier:  "xorm",
		TZLocation:     time.Local,
		tagHandlers:    defaultTagHandlers,
		cachers:        make(map[string]core.Cacher),
		defaultContext: context.Background(),
	}

	//根据驱动类型萱蕚时间格式
	if uri.DbType == core.SQLITE {
		engine.DatabaseTZ = time.UTC
	} else {
		engine.DatabaseTZ = time.Local
	}

	//设置log
	logger := NewSimpleLogger(os.Stdout)
	logger.SetLevel(core.LOG_INFO)
	engine.SetLogger(logger)
	engine.SetMapper(core.NewCacheMapper(new(core.SnakeMapper)))

	runtime.SetFinalizer(engine, close)

	return engine, nil
}

// NewEngineWithParams new a db manager with params. The params will be passed to dialect.
func NewEngineWithParams(driverName string, dataSourceName string, params map[string]string) (*Engine, error) {
	engine, err := NewEngine(driverName, dataSourceName)
	engine.dialect.SetParams(params)
	return engine, err
}

// Clone clone an engine
func (engine *Engine) Clone() (*Engine, error) {
	return NewEngine(engine.DriverName(), engine.DataSourceName())
}
