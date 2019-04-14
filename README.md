# nomof

`import "github.com/memememomo/nomof"`

nomof is a filter builder for guregu/dynamo.


```go
builder := NewBuilder()


// Case1

builder.Equal("Name", "Taro")
builder.Equal("Name", "Hanako")
builder.JoinAnd() // => ('Name' = ?) AND ('Name' = ?)
builder.Args      // => []string{"Taro", "Hanako"}


// Case2

builder.Op("Name1", EQ, "Taro1")
builder.Op("Name2", NE, "Taro2")
builder.Op("Name3", LT, "Taro3")
builder.Op("Name4", LE, "Taro4")
builder.Op("Name5", GT, "Taro5")
builder.Op("Name6", GE, "Taro6")
builder.JoinAnd() // => ('Name1' = ?) AND ('Name2' <> ?) AND ('Name3' < ?) AND ('Name4' <= ?) AND ('Name5' > ?) AND ('Name6' >= ?)
builder.Args()    // => []string{"Taro1","Taro2","Taro3","Taro4","Taro5","Taro6"}


// Case3

b1 := NewBuilder()
b2 := NewBuilder()
b1.Equal("Name", "Taro").Equal("Name", "Hanako")
b2.Equal("Age", 1)
b2.Append(b1.JoinOr(), b1.Arg)
b2.JoinAnd() // => ('Age' = ?) AND (('Name' = ?) OR ('Name' = ?))
b2.Args      // => []string{1, "Taro", "Hanako"}
```
