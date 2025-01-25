package testutil

import (
	"errors"
	"testing"
)

// go test -v -count=1 -timeout 60s

// go test -v -count=1 -timeout 60s -run ^TestGetFirst$
func TestGetFirst(t *testing.T) {
	t.Run("assert", func(t *testing.T) {
		AssertEqual(t, GetFirst(func() (string, error) { return "a", errors.New("") }()), "a")
		AssertEqual(t, GetFirst(func() (string, string, error) { return "a", "b", errors.New("") }()), "a")
		AssertEqual(t, GetFirst(func() (string, string, string, error) { return "a", "b", "c", errors.New("") }()), "a")
	})
}

// go test -v -count=1 -timeout 60s -run ^TestTest$
func TestTest(t *testing.T) {
	t.Run("AssertEqual", func(t *testing.T) {
		AssertEqual(t, true, true)
		AssertEqualFatal(t, true, true)
		AssertNotEqual(t, true, false)
		AssertNotEqualFatal(t, true, false)
		AssertEqual(t, GetLast(func() (string, string) { return "test1", "test2" }()), "test2")
	})
	t.Run("AssertContainStr", func(t *testing.T) {
		AssertContainStr(t, "ttest tst tttt", "test")
	})
	t.Run("AssertTrue", func(t *testing.T) {
		AssertTrue(t, true)
		AssertFalse(t, false)
	})
	t.Run("AssertDeepEqual", func(t *testing.T) {
		AssertDeepEqual(t, struct{ L []string }{L: []string{"a", "b", "c"}}, struct{ L []string }{L: []string{"a", "b", "c"}})
		AssertNotDeepEqual(t, struct{ L []string }{L: []string{"a", "b", "c"}}, struct{ L []string }{L: []string{"a", "b", "c", "d"}})
	})
	t.Run("AssertJsonExact", func(t *testing.T) {
		AssertJsonExact(t, `{"ID":"id","Name":"name"}`, `{"ID":"id","Name":"name"}`, nil)
		AssertJsonExact(t, `{"ID":null, "Name":null}`, `{"ID":null, "Name":null}`, []string{"ID"})
		AssertJsonExact(t,
			`{
			"result": "success",
			"id": "1234",
			"entity": {
			  "created_at": "2023-05-04T15:13:15.123456Z",
			  "name": "dummy"
			},
			"group": {
			  "member":{"number":3},
			  "team":{"name":"TEAM-A"},
			  "created_at": "2023-05-04T15:13:15.123456Z"
			}
		}`,
			`{
			"result": "",
			"id": "",
			"entity": {
			  "created_at": "",
			  "name": "dummy"
			},
			"group": {
			  "member":{"number":0},
			  "team":{"name":""},
			  "created_at": ""
			}
		}`,
			[]string{"result", "*.id", "*.created_at", "group"})

		AssertJsonExact(t,
			`{
				"bool": true,
				"float": 3.05,
				"int": 3,
				"list": [1,2,3,4,{"name":"taro"}],
				"nest": {
					"nest_nest":{
						"string":"string",
						"string2":"string2"
					},
					"members": [
						{
							"id":"3333",
							"created_at":"2023-05-04T15:13:15.123456Z"
						},
						{
							"id":"3334",
							"created_at":"2023-05-05T15:13:15.123456Z"
						}
					]
				}
			}`,
			`{
				"bool": false,
				"float": 0.0,
				"int": 0,
				"list": [0,0,0,0,{"name":""}],
				"nest": {
					"nest_nest":{
						"string":"string",
						"string2":""
					},
					"members": [
						{
							"id":"3333",
							"created_at":""
						},
						{
							"id":"3334",
							"created_at":""
						}
					]
				}
			}`,
			[]string{"bool", "float", "int", "notExistField", "nest.nest_nest.string2", "*.created_at", "list"})

		// 配列の場合にも対応した。(24/8/17)
		AssertJsonExact(t,
			`{
					"data": [{"name":"taro", "age":10},{"name":"jiro", "age":6}]
			}`,
			`{
					"data": [{"name":"", "age":10},{"name":"", "age":6}]
			}`,
			[]string{"data.name"})
	})

	t.Run("AssertComp", func(t *testing.T) {
		AssertComp(t, `{
				"bool": true,
				"float": 3.05,
				"int": 3,
				"list": [1,2,3,4,{"name":"taro"}],
				"nest": {
					"nest_nest":{
						"string":"string",
						"string2":"string2"
					},
					"members": [
						{
							"id":"3333",
							"created_at":"2023-05-04T15:13:15.123456Z"
						},
						{
							"id":"3334",
							"created_at":"2023-05-05T15:13:15.123456Z"
						}
					]
				}
			}`, `{
				"bool": true,
				"float": 3.05,
				"int": 3,
				"list": [1,2,3,4,{"name":"taro"}],
				"nest": {
					"nest_nest":{
						"string":"string",
						"string2":"string2"
					},
					"members": [
						{
							"id":"3333",
							"created_at":"2023-05-04T15:13:15.123456Z"
						},
						{
							"id":"3334",
							"created_at":"2023-05-05T15:13:15.123456Z"
						}
					]
				}
			}`)
	})
}
