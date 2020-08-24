# xorm源码学习

orm:对象关系映射是通过使用描述对象和数据库之间映射的元数据，将面向对象语言程序中的对象自动持久化到关系数据库中。



## orm对应关系

数据库                面向对象的编程语言
表(table)            类(class/struct)
记录(record, row)    对象 (object)
字段(field, column)    对象属性(attribute)

## xorm的结构
- xorm:xorm框架总入口
- engine:xorm内部核心,数据库管理器
- dialect:对象数据库类型转换关系对象
- session:和数据库的会话,暴露orm的语句
- statement:预创建sql
- cache:缓存模块
- tag:tag的handler(钩子)


## 调用层次
engine
    session
    dialect
        statement
            sql

- engine:
    - 暴露操作数据库orm的接口
- session:
    - 会话sql
- statement:
    - 预创建sql
    - 原生sql


- engine: 调用session的orm接口来达到crud的目的, 不关心底层
- session: 维护一次sql会话,可以链式调用, 通过底层的statement预缓存sql
- statement: 用于缓存上层调用接口的"sql半成品", 每次exec/query后都会清空


dialect 不同的数据库驱动, 有特定的实现:

- 数据类型转换
- 数据表检查, 索引检查
- 获取列
- 获取数据表
- 获取索引
- 创建表

除了数据类型转换, 各个驱动的不同点是元数据的存储地方(表)不同, 获取方式不同

只分析mysql, 其他都是大同小异

## 参考
- [geeorm](https://github.com/geektutu/7days-golang/tree/689a6d01b7dd04cd988faeb16a7f3041617873d2/gee-orm)
- [tinyorm](https://github.com/liangjfblue/tinyorm.git)