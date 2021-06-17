package algo

import (
	"fmt"
	"strings"
)

// JSONMap map[string]interface{}
type JSONMap map[string]interface{}

// JSONArray []interface{}
type JSONArray []interface{}

/***************** JSONMap的方法 ******************/

// GetString 根据key获取map里value为string类型的值
func (c JSONMap) GetString(key string) (string, error) {
	ret, ok := c[key].(string)
	if ok {
		return ret, nil
	}
	return "", fmt.Errorf("查找失败，key:%s", key)
}

// GetInt64 根据key获取map里value为int64类型的值
func (c JSONMap) GetInt64(key string) (int64, error) {
	ret, ok := c[key].(int)
	if ok {
		return int64(ret), nil
	}
	return 0, fmt.Errorf("查找失败，key:%s", key)
}

// GetInt32 根据key获取map里value为int32类型的值
func (c JSONMap) GetInt32(key string) (int32, error) {
	ret, ok := c[key].(int)
	if ok {
		return int32(ret), nil
	}
	return 0, fmt.Errorf("查找失败，key:%s", key)
}

// GetInt 根据key获取map里value为int类型的值
func (c JSONMap) GetInt(key string) (int, error) {
	ret, ok := c[key].(int)
	if ok {
		return ret, nil
	}
	return 0, fmt.Errorf("查找失败，key:%s", key)
}

// GetFloat64 根据key获取map里value为Float64类型的值
func (c JSONMap) GetFloat64(key string) (float64, error) {
	ret, ok := c[key].(float64)
	if ok {
		return ret, nil
	}
	return 0, fmt.Errorf("查找失败，key:%s", key)
}

// GetFloat32 根据key获取map里value为Float32类型的值
func (c JSONMap) GetFloat32(key string) (float32, error) {
	ret, ok := c[key].(float64)
	if ok {
		return float32(ret), nil
	}
	return 0, fmt.Errorf("查找失败，key:%s", key)
}

// GetMap 根据key获取map里value为JSONMap类型的值
func (c JSONMap) GetMap(key string) (JSONMap, error) {
	ret, ok := c[key].(JSONMap)
	if ok {
		return ret, nil
	}
	return nil, fmt.Errorf("查找失败，key:%s", key)
}

// GetArray 根据key获取map里value为JSONMap类型的值
func (c JSONMap) GetArray(key string) (JSONArray, error) {
	ret, ok := c[key].(JSONArray)
	if ok {
		return ret, nil
	}
	return nil, fmt.Errorf("查找失败，key:%s", key)
}

// GetStringByPath 根据key获取map里value为JSONMap类型的值
func (c JSONMap) GetStringByPath(path string) (string, error) {
	options := strings.Split(path, ".")
	var result string
	var middle JSONMap = c
	var err error
	for i, length := 0, len(options); i < length; i++ {
		if i == length-1 {
			result, err = middle.GetString(options[i])
		} else {
			middle, err = middle.GetMap(options[i])
		}
		if err != nil {
			return "", err
		}
	}
	return result, nil
}

// GetIntByPath 123
func (c JSONMap) GetIntByPath(path string) (int, error) {
	options := strings.Split(path, ".")
	var result int
	var middle JSONMap = c
	var err error
	for i, length := 0, len(options); i < length; i++ {
		if i == length-1 {
			result, err = middle.GetInt(options[i])
		} else {
			middle, err = middle.GetMap(options[i])
		}
		if err != nil {
			return 0, err
		}
	}
	return result, nil
}

// GetFloat64ByPath 123
func (c JSONMap) GetFloat64ByPath(path string) (float64, error) {
	options := strings.Split(path, ".")
	var result float64
	var middle JSONMap = c
	var err error
	for i, length := 0, len(options); i < length; i++ {
		if i == length-1 {
			result, err = middle.GetFloat64(options[i])
		} else {
			middle, err = middle.GetMap(options[i])
		}
		if err != nil {
			return 0, err
		}
	}
	return result, nil
}

/***************** JSONArray的方法 ******************/

// GetString 根据index获取map里value为string类型的值
func (c JSONArray) GetString(index int) (string, error) {
	ret, ok := c[index].(string)
	if ok {
		return ret, nil
	}
	return "", fmt.Errorf("查找失败，index:%d", index)
}

// GetInt 根据index获取map里value为int类型的值
func (c JSONArray) GetInt(index int) (int, error) {
	ret, ok := c[index].(int)
	if ok {
		return ret, nil
	}
	return 0, fmt.Errorf("查找失败，index:%d", index)
}

// GetFloat64 根据index获取map里value为Float64类型的值
func (c JSONArray) GetFloat64(index int) (float64, error) {
	ret, ok := c[index].(float64)
	if ok {
		return ret, nil
	}
	return 0, fmt.Errorf("查找失败，index:%d", index)
}

// GetFloat32 根据index获取map里value为Float32类型的值
func (c JSONArray) GetFloat32(index int) (float32, error) {
	ret, ok := c[index].(float64)
	if ok {
		return float32(ret), nil
	}
	return 0, fmt.Errorf("查找失败，index:%d", index)
}

// GetArray 根据index获取map里value为JSONArray类型的值
func (c JSONArray) GetArray(index int) (JSONArray, error) {
	ret, ok := c[index].(JSONArray)
	if ok {
		return ret, nil
	}
	return nil, fmt.Errorf("查找失败，index:%d", index)
}

// GetMap 根据index获取map里value为JSONMap类型的值
func (c JSONArray) GetMap(index int) (JSONMap, error) {
	ret, ok := c[index].(JSONMap)
	if ok {
		return ret, nil
	}
	return nil, fmt.Errorf("查找失败，index:%d", index)
}
