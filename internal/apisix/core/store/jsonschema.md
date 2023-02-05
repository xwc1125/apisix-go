# json schema

## 简介

json-schema，是用 json 的格式来定义 json 结构的方法，可以通过 json-schema 的定义规则，来检查 json 结构是否符合预期。

## 使用 `Json schema` 的优势

- 可描述的 json 数据格式
- 提供清晰的人工和机器可读文档
- 完整的 json schema 结构定义，更有利于测试及快速验证 json 的合法性

## `Json schema` 关键字说明

| 关键字                  | 描述                                                          |
|----------------------|-------------------------------------------------------------|
| $schema              | 	说明是哪个版本的 Json Schema，不同版本间不完全兼容                            |
| title                | 	用于进行简单的描述，可以省略                                             |
| description          | 	用于进行详细的描述信息，可以省略                                           |
| type                 | 	用于约束 json 的字段类型、string、number\integer、object、array、boolean |
| properties           | 	定义属性字段，定义每个字段的键和值类型                                        |
| required             | 	必需属性，数组类型，指定上述 properties 中哪些是必需的字段                        |
| minimum              | 	type 为 integer 时，可被接受的最小值                                  |
| maximum              | 	type 为 integer 时，可被接受的最大值                                  |
| maxLength            | 	type 为 string 或 array 时，可被接受的最小长度                          |
| minLength            | 	type 为 string 或 array 时，可被接受的最大长度                          |
| uniqueItems          | 	type 为 array 时，数组元素时否唯一                                    |
| pattern              | 	type 为 string 时，将字符串限制为特定的正则表达式                            |
| enum                 | 	用于将值限制为一组固定的值，其至少含有一个元素，且所含元素必须唯一                          |
| additionalProperties | 	将用于验证与 properties 不匹配的属性是否被允许                              |
| patternProperties    | 	将正则表达式映射到模式。如果属性名称与给定的正则表达式匹配，则属性值必须符合定义的类型                |
| allOf                | 	(AND) 必须对「所有」子模式有效                                         |
| anyOf                | 	(OR) 必须对「任意」子模式有效                                          |
| oneOf                | 	(XOR) 必须对「恰好一个」子模式有效                                       |

## 参考文献

- https://json-schema.apifox.cn/