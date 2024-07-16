package interpreter

import (
	"encoding/json"
	"fmt"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/stretchr/testify/assert"
	url2 "net/url"
	"testing"
)

func get[T any](a any) (ans T, f bool) {
	if val, ok := a.(T); ok {
		ans = val
		f = true
	}
	return
}

func TestBadScriptToNewInterpreter(t *testing.T) {
	log.Setup(&log.LoggerConfig{
		Debug:      true,
		CallerSkip: 1,
	})
	_, err := NewEvalInterpreter(`
func Check(data map[string]an) string {
	return data["name"].(string)
}
	`)
	assert.Error(t, err)
}

func TestNewEvalInterpreter(t *testing.T) {
	evalInterpreter, err := NewEvalInterpreter(`
func Check(data map[string]any) string {
	return data["name"].(string)
}
	`)
	assert.Equal(t, err, nil)
	s := "{\n \"name\" : \"john\",\n\"id\": 1\n}"
	var data map[string]any
	err = json.Unmarshal([]byte(s), &data)
	assert.Equal(t, err, nil)
	var result string
	err = evalInterpreter.ExecuteScript(func(executor any) {
		result = executor.(func(map[string]any) string)(data)
	})
	assert.Equal(t, err, nil)
	assert.Equal(t, "john", result)
}

func TestNewEvalInterpreterJudge(t *testing.T) {
	evalInterpreter, err := NewEvalInterpreter(`
import "fmt"
func Check(data map[string]any) string {
	s, ok := get[string](data["name"])
	fmt.Println(s)
    if ok && s == "john"{
		return "good"
	}
	return "fuck"
}
	`)
	if err != nil {
		t.Error(err)
	}
	s := "{\n \"name\" : \"john\",\n\"id\": 1\n}"
	var data map[string]any
	err = json.Unmarshal([]byte(s), &data)
	assert.Equal(t, err, nil)
	var result string
	err = evalInterpreter.ExecuteScript(func(executor any) {
		result = executor.(func(map[string]any) string)(data)
	})
	assert.Equal(t, "good", result)
	assert.Equal(t, err, nil)
}

func TestNewEvalInterpreterWithFunc(t *testing.T) {
	evalInterpreter, err := NewEvalInterpreter(`
func Check2(data map[string]any) string {
	s, ok := get[string](data["name"])
    if ok && s == "john"{
		return "good"
	}
	return "fuck"
}
	`, WithFuncName("Check2"))
	if err != nil {
		t.Error(err)
	}
	s := "{\n \"name\" : \"john\",\n\"id\": 1\n}"
	var data map[string]any
	err = json.Unmarshal([]byte(s), &data)
	assert.Equal(t, err, nil)
	var result string
	err = evalInterpreter.ExecuteScript(func(executor any) {
		result = executor.(func(map[string]any) string)(data)
	})
	assert.Equal(t, "good", result)
	assert.Equal(t, err, nil)
}

func TestNewEvalInterpreterWithCheckFunc(t *testing.T) {
	evalInterpreter, err := NewEvalInterpreter(`
import (
	"encoding/json"
	"fmt"
	"strconv"
)
func Check(responseBody string) string {
	var data map[string]any
	err := json.Unmarshal([]byte(responseBody), &data)
	if err != nil {
		return "error parse"
	}
    // if the response body has code key, and the value is 200, return good
	if iCode, err := strconv.Atoi(fmt.Sprintf("%v", data["code"])); err == nil && iCode == 200 {
		return "good"
	}
    // else return bad
	return "bad"
}
	`, WithFuncName("Check"))
	if err != nil {
		t.Error(err)
	}
	s := "{\n \"code\" : 200 ,\n\"id\": 1\n}"
	assert.Equal(t, err, nil)
	var result string
	err = evalInterpreter.ExecuteScript(func(executor any) {
		result = executor.(func(string) string)(s)
	})
	assert.Equal(t, "good", result)
	assert.Equal(t, err, nil)
}

func TestNewEvalInterpreterWithDynamic(t *testing.T) {
	evalInterpreter, err := NewEvalInterpreter(`
import "encoding/json"

func RndParam() (data []map[string]map[string]any) {
	data = append(data, map[string]map[string]any{
		"header": {
			"Content-Type": "application/json",
		},
		"query": {
			"id": 0,
			"alias": "bob0",
		},
		"body": {
			"name": "john",
			"info": map[string]any{
				"age": "18",
				"sex": "male",
			},
		},
	})
	data = append(data, map[string]map[string]any{
		"header": {
			"Content-Type": "application/json",
		},
		"query": {
			"id": 1,
			"alias": "bob1",
		},
		"body": {
			"name": "john",
			"info": map[string]any{
				"age": "18",
				"sex": "male",
			},
		},
	})
	s := "{\n    \"header\": {\n        \"Content-Type\": \"application/json\"\n    },\n    \"query\": {\n        \"id\": 2,\n \"alias\": \"bob2\"\n    },\n    \"body\": {\n        \"name\": \"john\",\n        \"info\": {\n            \"age\": \"18\",\n            \"sex\": \"male\"\n        }\n    }\n}"
	var data3 map[string]map[string]any
	json.Unmarshal([]byte(s), &data3)
	data = append(data, data3)
	return data
}
		`, WithFuncName("RndParam"))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, err, nil)
	var dataList []map[string]map[string]any
	err = evalInterpreter.ExecuteScript(func(executor any) {
		dataList = executor.(func() []map[string]map[string]any)()
	})
	assert.Equal(t, err, nil)
	for _, data := range dataList {
		values := url2.Values{}
		for key, value := range data["query"] {
			values.Add(key, fmt.Sprintf("%v", value))
		}
		t.Log(values.Encode())
		assert.Equal(t, data["header"]["Content-Type"], "application/json")
		assert.Equal(t, data["body"]["name"], "john")
		assert.Equal(t, data["body"]["info"].(map[string]any)["age"], "18")
		assert.Equal(t, data["body"]["info"].(map[string]any)["sex"], "male")
	}
}

func TestUrlParse(t *testing.T) {
	url, err := url2.Parse("https://www.google.com?key1=value1&key2=value2")
	assert.Equal(t, err, nil)
	println(url.RawQuery)
	println(fmt.Sprintf("%v", 2.2))
}

// todo benchmark test eval script by interpreter
