package common

import "reflect"

// StructToMap 将结构体转换为 map
func StructToMap(obj interface{}) map[string]interface{} {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	objType := objValue.Type()

	data := make(map[string]interface{})

	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldName := objType.Field(i).Name
		fieldValue := field.Interface()

		if field.Type().Kind() == reflect.Struct {
			if field.Kind() == reflect.Ptr && !field.IsNil() {
				// 如果字段是指针类型的结构体且非空指针，则获取指针指向的值
				fieldValue = StructToMap(field.Elem().Interface())
			} else {
				// 否则，递归处理嵌套的结构体
				fieldValue = StructToMap(fieldValue)
			}
		}

		data[fieldName] = fieldValue
	}

	return data
}
