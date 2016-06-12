package main

import (
    "fmt"
    "reflect"
)

func main() {
	data := map[string]interface{} {
		"email": "huangxin@tenxcloud.com",
		"phone": "13811444541",
		"groupId": 2,
	}
	email := getVal(data, "email", "")
	fmt.Println("get", email)
}

func getVal(data map[string]interface{}, key string, 
	        defaultVal interface{}) (val interface{}) {
	val, ok := data[key]
    if !ok || reflect.TypeOf(val) != reflect.TypeOf(defaultVal) {
    	fmt.Printf("cannot get %T value for param '%s' \n", defaultVal, key)
    	val = defaultVal
	}
	return 
}
