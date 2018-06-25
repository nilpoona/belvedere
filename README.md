# belvedere


## Usage


Prepare a structure that defines table information.
```go

// Definition of `user` table.
type User struct {
  ID        uint64 `pk:"true"` // Defines that the id column is primary.
  Name      string
  Age       uint
  Gendor    string
}
```

Connect to database.
```go
b, e := NewBelvedere("mysql", "test:test@/test?parseTime=true")
if e != nil {
  // handle error.
}
```

Create record.
```go
b, e := NewBelvedere("mysql", "test:test@/test?parseTime=true")
if e != nil {
  // handle error.
}
u := &User{
  ID: 1,
  Name: "foo",
  Age: 22,
  Gendor: "male",
}

// Insert record in `user` table.
r, e := b.Insert(ctx, u)
if e != nil {
  // handle error.
}
```

Get record.
```go
b, e := NewBelvedere("mysql", "test:test@/test?parseTime=true")
if e != nil {
  // handle error.
}

u := &User{
  ID: 1,
}

// Retrieve record with condition of primary key value.
r, e := b.SelectOne(ctx, u)
if e != nil {
  // handle error.
}

// Display value of name column.
fmt.Println(u.Name)
```

```go

// Retrieve the record that satisfies the condition.
var users []*User

// Acquire users over 20 years old.
e := b.Select(ctx, &users, Where("age > ?", 20))

// Get male users over 20 years old.
e := b.Select(
  ctx,
  &users,
  And(Where("age > ?", 20), Where("gendor > ?", "male")),
 )

// 10 male users over the age of 20 are acquired.
e := b.Select(
  ctx,
  &users,
  And(Where("age > ?", 20), Where("gendor > ?", "male")),
  Limit(10),
 )

```
