package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

/**
用source的所有字段覆盖target的
如果fields不为空, 表示用source的特定字段覆盖target的
target应该为结构体指针
**/

func CopyFields(source interface{}, target interface{}, fields ...string) (err error) {

	targetType := reflect.TypeOf(target)
	targetValueArray := reflect.ValueOf(target)
	sourceType := reflect.TypeOf(source)
	sourceValueArray := reflect.ValueOf(source)

	if source == nil {
		return nil
	}

	// 简单判断下，目标类型必须为指针,原类型必须为结构体
	if targetType.Kind() != reflect.Ptr {
		return errors.New("target must be a struct pointer")
	}
	if sourceType.Kind() != reflect.Struct {
		return errors.New("source must be a struct")
	}

	targetValueArray = reflect.ValueOf(targetValueArray.Interface())
	// 要复制哪些字段
	_fields := make([]string, 0)
	if len(fields) > 0 {
		_fields = fields
	} else {
		for i := 0; i < sourceValueArray.NumField(); i++ {
			_fields = append(_fields, sourceType.Field(i).Name)
		}
	}
	if len(_fields) == 0 {
		return errors.New("no fields to copy")
	}
	// 复制
	for i := 0; i < len(_fields); i++ {
		name := _fields[i]
		targetValue := targetValueArray.Elem().FieldByName(name)
		sourceValue := sourceValueArray.FieldByName(name)
		// target中有同名的字段并且类型一致才复制
		if targetValue.IsValid() && targetValue.CanSet() && targetValue.Kind() == sourceValue.Kind() {
			targetKind := targetValue.Kind()
			sourceKind := sourceValue.Kind()
			if sourceValue.Kind() == reflect.Slice {
				//targetSliceValueArray := reflect.ValueOf(targetValue)
				for j := 0; j < sourceValue.Len(); j++ {
					sourceSliceValue := sourceValue.Index(j)
					kind := sourceSliceValue.Kind()
					fmt.Println(kind)
				}
			}
			fmt.Println(targetKind, sourceKind)
			targetValue.Set(sourceValue)
		}
		//else {
		//	logTracing.LogPrintfP("no such field or different kind, fieldName: %s\n", name)
		//}
	}
	return
}

func CopyFieldsByJson(source interface{}, target interface{}) (err error) {
	// 将原始结构体转换为JSON格式的字节数组
	b, err := json.Marshal(source)
	if err != nil {
		return err
	}
	// 将JSON格式的字节数组解析成目标结构体
	err = json.Unmarshal(b, target)
	if err != nil {
		return err
	}
	return nil
}

// FillOmitFields 补充数据
func FillOmitFields(s interface{}) {
	// 获取结构体的类型
	t := reflect.TypeOf(s)

	// 判断类型是否为指针类型
	if t.Kind() == reflect.Ptr {
		// 获取指针指向的类型
		t = t.Elem()
	}

	// 遍历结构体的所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 判断字段是否带有omitempty标签
		if _, ok := field.Tag.Lookup("omitempty"); ok {

			// 判断字段是否为空
			fieldValue := reflect.ValueOf(s).Elem().Field(i)
			if fieldValue.IsZero() {

				// 根据字段类型设置零值
				switch fieldValue.Kind() {
				case reflect.String:
					fieldValue.SetString("")
				case reflect.Int:
					fieldValue.SetInt(0)
				case reflect.Bool:
					fieldValue.SetBool(false)
					// ... 其他类型
				}
			}
		}
	}
}
