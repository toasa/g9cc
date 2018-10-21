package util

import (
    "testing"
    . "g9cc/common"
)

func TestVector(t *testing.T) {
    tests := []struct {
        index int
        expectedValue int
    }{
        {5, 5},
        {50, 50},
        {99, 99},
    }

    var vec *Vector = New_vec()
    for i := 0; i < 100; i++ {
        Vec_push(vec, i)
    }

    for _, tt := range tests {
        d, _ := vec.Data[tt.index].(int)

        if d != tt.expectedValue {
            t.Fatalf("%d expected, but got %d", tt.expectedValue, d)
        }
    }
}

func TestMap(t *testing.T) {
    tests := []struct {
        key string
        putValue interface{}
        expectedValue interface{}
    }{
        {"foo", nil, 0},
        {"foo", 2, 2},
        {"bar", 4, 4},
        {"foo", 6, 6},
    }

    var m *Map = New_map()

    for _, tt := range tests {
        Map_put(m, tt.key, tt.putValue)
        mapValue, _ := Map_get(m, tt.key).(int)
        getValue, _ := tt.expectedValue.(int)
        if mapValue != getValue {
            t.Fatalf("%d expected, but got %d", getValue, mapValue)
        }
    }
}
