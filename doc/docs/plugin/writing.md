# 编写插件

给`盘古`编写插件非常非常容易，只需要简单的两个步骤就能完成

- 编写功能方法（建议用`Struct`进行封装）
- 添加构造函数
- 在初始化方法中完成组件的注入

## 功能

<<< @/../example/plugin/test.go{13}

注意

- 功能函数必须是公开的（除非是插件内部使用）
- 接收者可以是指针也可以不是指针
- 强烈建议使用`Struct`封装内部实现，尽量不对外暴露实现细节

## 构造函数

<<< @/../example/plugin/test.go{9}

注意

- 构造函数可以是公开也可以是私有构造函数，取决于开发者想不想暴露给用户直接使用
- 建议私有构造函数优先

## 注入

要接入`盘古`框架，只需要在包的初始化方法`init`中，注入功能组件到`盘古`框架中即可

<<< @/../example/plugin/init.go

其中

- `newAAA`、`newBBB`以及`newCCC`按普通构造函数编写代码即可
- 如果在初始化过程中出现错误（`error`不为空），建议立即抛出异常