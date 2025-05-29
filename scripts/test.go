// go:build ignore

package main

func main() {
	arr := [...]int{1, 2, 3, 4, 5}

	printArr(arr[3:])
}

func printArr(arr []int) {
	for i := 0; i < len(arr); i++ {
		println(arr[i])
	}
}
