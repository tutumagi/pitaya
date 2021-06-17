package algo

import (
	"testing"
)

func Test_JsonMap(t *testing.T) {
	data := JSONMap{
		"arrayData": JSONArray{
			"abc",
			JSONMap{
				"number1": 1,
			},
		},
		"number2": 2,
		"string1": "hao",
		"string": JSONMap{
			"demo":  1,
			"demo1": 1.2,
		},
	}

	if string1, err := data.GetString("string1"); err == nil {
		if string1 != "hao" {
			t.Errorf("should be hao, but is %v", string1)
		}
	} else {
		t.Errorf("should be hao, but is err: %v", err)
	}

	if number2, err := data.GetInt("number2"); err == nil {
		if number2 != 2 {
			t.Errorf("should be 2, but is %v", number2)
		}
	} else {
		t.Errorf("should be 2, but is err: %v", err)
	}

	if arrayData, err := data.GetArray("arrayData"); err == nil {
		if stringdata, err := arrayData.GetString(0); err == nil {
			if stringdata != "abc" {
				t.Errorf("should be abc, but is %v", stringdata)
			}
		} else {
			t.Errorf("should be abc, but is err %v", err)
		}

		if mapData, err := arrayData.GetMap(1); err == nil {
			if number1, err := mapData.GetInt("number1"); err == nil {
				if number1 != 1 {
					t.Errorf("should be 1, but is %v", number1)
				}
			}
		} else {
			t.Errorf("should be mapdata, but is err %v", err)
		}
	}

	if intData, err := data.GetIntByPath("string.demo"); err == nil {

		if intData != 1 {
			t.Errorf("should be 1, but is %v", intData)
		}
	} else {
		t.Errorf("should be 1, but is err %v", err)
	}

	if Data, err := data.GetFloat64ByPath("string.demo1"); err == nil {

		if Data != 1.2 {
			t.Errorf("should be 1.2, but is %v", Data)
		}
	} else {
		t.Errorf("should be 1.2, but is err %v", err)
	}

}

func Test_JsonArray(t *testing.T) {
	data := JSONArray{
		JSONMap{
			"demo":  1,
			"demo1": 1.2,
		},
		1,
		2.5,
		"HELLO",
		JSONArray{
			JSONMap{
				"demo":  1,
				"demo1": 1.2,
			},
			1,
			2.5,
			"HELLO",
		},
	}
	if Data, err := data.GetArray(4); err == nil {
		if data, err := Data.GetInt(1); err == nil {
			if data != 1 {
				t.Errorf("should be 1, but is: %v", data)
			}
		}
	} else {
		t.Errorf("should be exist, but is err: %v", err)
	}

}
