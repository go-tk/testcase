package testcase

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func (tc *TestCase) Dump(buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "Function Type: %v\n", tc.functionType)
	{
		fmt.Fprint(buffer, "Callback IDs: ")
		callbackIDStrs := make([]string, len(tc.callbackValues))
		i := 0
		for callbackID := range tc.callbackValues {
			callbackIDStrs[i] = fmt.Sprintf("%v", callbackID)
			i++
		}
		sort.Strings(callbackIDStrs)
		fmt.Fprint(buffer, strings.Join(callbackIDStrs, ", "))
		fmt.Fprint(buffer, "\n")
	}
}

func (tc *TestCase) DumpAsString() string {
	var buffer bytes.Buffer
	tc.Dump(&buffer)
	return buffer.String()
}
