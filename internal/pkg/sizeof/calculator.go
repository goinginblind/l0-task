/*
sizeof package is made primarly to calculate the sum of deep and shallow sizes of a struct.
It's intended use is to make the caching a bit more robust and allow to consider not only
the amount of cache entries, but also the memory they takeup.
*/
package sizeof

import "reflect"

// SizeOf calculates the approximate memory usage of a variable (shallow+deep data sizes)
// It traverses structs, slices, maps, pointers, and strings to account for heap-allocated data
func SizeOf(v any) int {
	// visited map is used to handle pointers and avoid infinite loops with circular references
	visited := make(map[uintptr]bool)
	return sizeOf(reflect.ValueOf(v), visited)
}

func sizeOf(v reflect.Value, visited map[uintptr]bool) int {
	if !v.IsValid() {
		return 0
	}
	if v.CanAddr() {
		ptr := v.UnsafeAddr()
		if visited[ptr] {
			return 0
		}
		visited[ptr] = true
	}

	var size int
	size += int(v.Type().Size())
	size += sizeOfHeapData(v, visited)

	return size
}

// sizeOfHeapData calculates the size of data that is not stored inline with the value itself.
// This applies to strings, slices, maps, and pointers.
func sizeOfHeapData(v reflect.Value, visited map[uintptr]bool) int {
	if !v.IsValid() {
		return 0
	}

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			return sizeOf(v.Elem(), visited)
		}
	case reflect.String:
		return v.Len()
	case reflect.Slice:
		if !v.IsNil() {
			size := v.Cap() * int(v.Type().Elem().Size())
			for i := 0; i < v.Len(); i++ {
				size += sizeOfHeapData(v.Index(i), visited)
			}
			return size
		}
	case reflect.Struct:
		size := 0
		for i := 0; i < v.NumField(); i++ {
			size += sizeOfHeapData(v.Field(i), visited)
		}
		return size
	case reflect.Map:
		if !v.IsNil() {
			size := 0
			for _, key := range v.MapKeys() {
				size += sizeOf(key, visited)
				size += sizeOf(v.MapIndex(key), visited)
			}
			return size
		}
	}

	return 0
}
