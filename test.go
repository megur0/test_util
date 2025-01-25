package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func GetFirst[A any](a A, b ...any) A {
	return a
}

func GetLast(a ...any) any {
	return a[len(a)-1]
}

func AssertEqual(t *testing.T, target any, expect any) {
	t.Helper()
	if target != expect {
		t.Errorf("%#v is not equal to %#v\n", target, expect)
	}
}

func AssertTrue(t *testing.T, target any) {
	t.Helper()
	AssertEqual(t, target, true)
}

func AssertFalse(t *testing.T, target any) {
	t.Helper()
	AssertEqual(t, target, false)
}

func AssertUnTypedNil(t *testing.T, target any) {
	t.Helper()
	AssertEqual(t, target, nil)
}

func AssertTypedNil[T any](t *testing.T, target *T) {
	t.Helper()
	AssertEqual(t, target, (*T)(nil))
}

func AssertNotUnTypedNil(t *testing.T, target any) {
	t.Helper()
	AssertNotEqual(t, target, nil)
}

func AssertNotTypedNil[T any](t *testing.T, target *T) {
	t.Helper()
	AssertNotEqual(t, target, (*T)(nil))
}

func AssertErrorAs(t *testing.T, target error, expect any) {
	t.Helper()
	if !errors.As(target, expect) {
		t.Errorf("error %#v is not as %#v\n", target, expect)
	}
}

func AssertDeepEqual(t *testing.T, target any, expect any) {
	t.Helper()
	if !reflect.DeepEqual(target, expect) {
		t.Errorf("%#v is not deep equal to %#v\n", target, expect)
	}
}

func AssertNotDeepEqual(t *testing.T, target any, expect any) {
	t.Helper()
	if reflect.DeepEqual(target, expect) {
		t.Errorf("unexpctedlly, %v is equal to %v\n", target, expect)
	}
}

func AssertEqualFatal(t *testing.T, target any, expect any) {
	t.Helper()
	if target != expect {
		t.Fatalf("%v is not equal to %v\n", target, expect)
	}
}

func AssertNotEqual(t *testing.T, target any, shouldNotBe any) {
	t.Helper()
	if target == shouldNotBe {
		t.Errorf("unexpectedly, %v is equal to %v\n", target, shouldNotBe)
	}
}

func AssertNotEqualFatal(t *testing.T, target any, shouldNotBe any) {
	t.Helper()
	if target == shouldNotBe {
		t.Fatalf("unexpectedly, %v is equal to %v\n", target, shouldNotBe)
	}
}

func AssertContainStr(t *testing.T, target any, str string) {
	t.Helper()
	targetStr, ok := target.(string)
	if !ok {
		t.Fatal("target is not string")
	}
	if !strings.Contains(targetStr, str) {
		t.Errorf("%v does not contai %s\n", target, str)
	}
}

func AssertComp(t *testing.T, target string, expect string) {
	t.Helper()
	if diff := cmp.Diff(expect, target); diff != "" {
		t.Errorf("Compare value is mismatch (-v1 +v2):%s\n", diff)
	}
}

// 文字列のjsonを受け取ってmapにしてDiffを取る。
// 完全一致比較。（余計なフィールドがあってもエラーとなり、不足してもエラーとなる）
//
// 現状、最上位が{}で囲まれているJSONのみに対応している
// これは、json.Unmarshalした際に、最上位がmap[string]interface{}となるもので、
// 例えば、[0,1,2,3] や 3 はエラーとなる。
//
// ignoreValuesで指定したフィールドはデフォルト値とすることができる。
// ignoreValuesはドット区切りで指定する。（例）data.entity.user.id
// したがってフィールドが存在することをチェックしたいが値自体には関心がない場合に、
// ignoreValuesを指定した上で、expectJsonの対象のフィールドを
// stringなら"", intなら0, boolならfalse, に指定すれば良い。
//
// (24/8/17対応)ignoreValuesは途中が配列の場合は全ての子要素に適用される。
// 例えば"data.id"のように指定した際に、{"data": [{"id":1}, {"id":2}, ...]}の全ての子要素に
// 適用される。ただし{"data": [[ {"id":1} ]] }のように入れ子の配列には適用されない。
//
// ブロックを指定（例えば、data.entity.userとした際にuser内に各種キーが含まれる場合）した場合は、その中身が再帰的に全てデフォルト値になる。
// (user内のプロパティが消えて空のマップとなるわけではなく、中のキーそれぞれがデフォルト値になる点に注意。)
// *.fieldNameのように指定すると階層に関係なくすべての"fieldName"
// をデフォルト値にする。（ただ、mapを全走査するので効率は悪い。）
// *は必ず先頭に置いて直後にフィールドを指定する必要がある。
// aaa.*.fieldNameや、*.aaa.fieldNameといった指定はできない。
//
// 日時はJSON上は単なる文字列のため、ゼロバリューは空文字となる。
func AssertJsonExact(t *testing.T, targetJson string, expectJson string, ignoreValues []string) {
	t.Helper()
	var expectedMap interface{}
	if err := json.Unmarshal([]byte(expectJson), &expectedMap); err != nil {
		t.Fatalf("AssertJson Unmarshal failed:%s \njson:\n%s", err, expectJson)
	}

	var targetMap interface{}
	if err := json.Unmarshal([]byte(targetJson), &targetMap); err != nil {
		t.Fatalf("AssertJson Unmarshal failed:%s \njson:\n%s", err, targetJson)
	}

	// ignoreで指定されたフィールドを再帰的にゼロバリューにする
	wildIgnoreField := []string{}
	for _, ignoreValueStr := range ignoreValues {
		m := targetMap.(map[string]interface{})
		ignoreField := strings.Split(ignoreValueStr, ".")

		// * はワイルドカードとして後続で別で処理。
		if ignoreField[0] == "*" {
			wildIgnoreField = append(wildIgnoreField, ignoreField[1])
			continue
		}

		resetIgnoreField(m, ignoreField)
	}

	// 「*.fieldName」のように指定されたフィールドはすべて値をゼロバリューにする
	if len(wildIgnoreField) != 0 {
		resetFieldsOnMap(targetMap.(map[string]interface{}), wildIgnoreField)
	}

	if diff := cmp.Diff(expectedMap, targetMap); diff != "" {
		t.Errorf("Compare value is mismatch (-v1 +v2):%s\n", diff)
	}
}

// 指定されたignoreFieldを探索して、ゼロバリューとする。
// ignoreで指定したキーが存在しない場合は、何もしない。
func resetIgnoreField(m map[string]interface{}, ignoreField []string) {
	field := ignoreField[0]
	if _, ok := m[field]; !ok {
		// キーが存在しない場合はここで終了する
		return
	}

	// 例えばもともとの指定された文字列が「"aaa.bbb.ccc.ddd"」だとすると、
	// この分岐では、ignoreFieldが["ddd"]となった状態となる。
	if len(ignoreField) == 1 {
		resetFieldOnMap(m, field)
		return
	}

	switch v := m[field].(type) {
	case []interface{}:
		// 例えば{"data":[{"xxx":yyy}, {"xxx":zzz}]}のような配列データは、
		// ignoreFieldが data.xxxの場合は 全ての子に適用する。
		for _, ve := range v {
			switch vv := ve.(type) {
			case map[string]interface{}:
				resetIgnoreField(vv, ignoreField[1:])
			case []interface{}:
				// 例えば{"data":[[{"xxx":yyy}]]}のような入れ子になった配列データ
				// は対象外とする。
				continue
			default:
				// その他の値は無視
				continue
			}
		}
		return
	case map[string]interface{}:
		resetIgnoreField(v, ignoreField[1:])
	default:
		// 値が複合型では無い場合はここで終了する
		return
	}
}

// map内の、fieldsで指定されたキーの値を再起的にゼロバリューにする
// 指定のフィールドが配列やマップの場合は、内部を再起的にゼロバリューにする。
func resetFieldsOnMap(m map[string]interface{}, fields []string) {
	for key, val := range m {
		if listContains(fields, key) {
			resetFieldOnMap(m, key)
		} else {
			switch vv := val.(type) {
			case string, bool, float64:
				continue
			case []interface{}:
				for _, ve := range vv {
					switch vvv := ve.(type) {
					case map[string]interface{}:
						resetFieldsOnMap(vvv, fields)
					case string, bool, float64:
						continue
					default:
						panic("")
					}
				}
			case map[string]interface{}:
				resetFieldsOnMap(vv, fields)
			default:
				if vv != nil {
					panic(fmt.Sprintf("unexpected type: %T", vv))
				}
				// nilの場合は特に何もしない。
				continue
			}
		}
	}
}

// map内の全てのフィールドを再帰的にゼロバリューにする
func resetMap(m map[string]interface{}) {
	for key := range m {
		resetFieldOnMap(m, key)
	}
}

// map内の指定のフィールド(key)の値をゼロバリューにする
// 指定のフィールドが配列やマップの場合は、内部を再起的にゼロバリューにする。
func resetFieldOnMap(m map[string]interface{}, key string) {
	switch vv := m[key].(type) {
	case string, bool, float64:
		m[key] = getZeroValueFromJsonField(vv)
	case []interface{}:
		for i, ve := range vv {
			switch vvv := ve.(type) {
			case map[string]interface{}:
				resetMap(vvv)
			case string, bool, float64:
				vv[i] = getZeroValueFromJsonField(vvv)
			default:
				panic("")
			}
		}
	case map[string]interface{}:
		resetMap(vv)
	default:
		if vv != nil {
			panic(fmt.Sprintf("unexpected type: %T", vv))
		}
		// nilの場合は特に何もしない。
		return
	}
}

// Unmarshalで取得したJsonの各フィールドのゼロバリューを返す。
func getZeroValueFromJsonField(val any) any {
	switch vv := val.(type) {
	// case bool, ...: という感じにまとめてしまうと、vv はanyになってしまうので
	// caseは分けている。
	case bool:
		return getZeroVal(vv)
	case float64: // 整数でもUnmarshalでfloat64に変換されるため、intやfloat32は無い。
		return getZeroVal(vv)
	case string:
		return getZeroVal(vv)
	default:
		panic("")
	}
}

func getZeroVal[R any](_ R) R {
	return *new(R)
}

func listContains[T comparable](slice []T, val T) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
