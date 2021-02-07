package resp

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	crByte              = byte('\r')
	nlByte              = byte('\n')
	whitespaceByte      = byte(' ')
	stringStartByte     = byte('+')
	integerStartByte    = byte(':')
	bulkStringStartByte = byte('$')
	arrayStartByte      = byte('*')
	errorStartByte      = byte('-')
)

// Note: This is not the REDIS standard error codes.
//Just using to notify to client about some errors description
const (
	InvalidByteSeq = "IVBYTESEQ"
)

// Assert a non-empty byte stream or panic otherwise
func assertNonEmptyStream(bytes []byte) {
	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}
}

// Assert start symbol of a byte stream (start byte) matches expected start byte (symbol)
func assertStartSymbol(startByte byte, symbol byte) {
	if startByte != symbol {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", symbol, startByte)))
	}
}

// Utility function to read a byte stream until CRLF
// and return the number of bytes consumed along with read bytes.
func readUntilCRLF(bytes []byte, excludeFirstByte bool) (string, int) {
	var sb strings.Builder
	var c byte
	read := 0
	i := 0
	if excludeFirstByte == true && len(bytes) > 0 {
		i = 1
		read = 1
	}
	// Build string
	for ; i < len(bytes); i++ {
		c = bytes[i]
		read++
		if c == nlByte {
			break
		}
		if c != crByte {
			sb.WriteByte(c)
		}
	}
	return sb.String(), read
}

// Parse a simple string from bytes and return parsed string and number of bytes consumed
func parseSimpleString(bytes []byte) (String, int) {
	assertNonEmptyStream(bytes)
	assertStartSymbol(bytes[0], stringStartByte)
	str, i := readUntilCRLF(bytes, true)
	// Return value and bytes read
	return NewString(str), i
}

// Parse a sequence of bytes as per Integer specification.
func parseIntegers(bytes []byte) (Integer, int) {
	assertNonEmptyStream(bytes)
	assertStartSymbol(bytes[0], integerStartByte)
	str, i := readUntilCRLF(bytes, true)
	// Return value and bytes read
	conv, err := strconv.Atoi(str)
	if err != nil {
		panic(fmt.Sprintf("Invalid integer sequence supplied: %s", str))
	}
	return NewInteger(conv), i
}

// parse a sequence of bytes representing bulk string
func parseBulkString(bytes []byte) (BulkString, int) {
	assertNonEmptyStream(bytes)
	assertStartSymbol(bytes[0], bulkStringStartByte)
	// Convert start symbol to : and parse as integer
	bytes[0] = integerStartByte
	str := ""
	isNullValue := false
	respInt, read := parseIntegers(bytes)
	read2 := 0
	// This check is much faster than the length check in constructor.
	// It is safer to fail here.
	if respInt.GetIntegerValue() > (MaxBulkSizeLength) {
		panic("Bulk string length exceeds maximum allowed size of " + MaxBulkSizeAsHumanReadableValue)
	} else if respInt.GetIntegerValue() < -1 {
		panic("Bulk string length must be greater than -1")
	} else {
		switch respInt.GetIntegerValue() {
		case 0:
			// Short circuit
			break
		case -1:
			// Null string
			isNullValue = true
			break
		default:
			// Regular parse
			bytes = bytes[read:]
			str, read2 = readUntilCRLF(bytes, false)
			if len(str) != respInt.GetIntegerValue() {
				panic(fmt.Sprintf("Bulk string length %d does not match expected length of %d", len(str), respInt.GetIntegerValue()))
			}
			break
		}
	}
	if isNullValue {
		return NewNullBulkString(), read + read2
	}
	bs, err := NewBulkString(str)
	if err != nil {
		panic(err)
	}
	return bs, read + read2
}

// parseArray parses a sequence of bytes as per RESP array specifications.
// Clients typically send commands as Array
func parseArray(bytes []byte) (*Array, int) {
	assertNonEmptyStream(bytes)
	assertStartSymbol(bytes[0], arrayStartByte)
	bytesRead := 0
	// Reset to colon so we can parse integer
	bytes[0] = integerStartByte
	//Get Number of elements in array
	numberOfItems, n := parseIntegers(bytes)
	bytesRead += n
	// Create new Array
	Array, err := NewArray(numberOfItems.GetIntegerValue())
	if err != nil {
		panic(err)
	}
	counter := 0
	// Advance bytes
	bytes = bytes[n:]
	var s IDataType
	var r int
	for len(bytes) > 0 {
		if counter >= Array.GetNumberOfItems() {
			panic(fmt.Sprintf("Invalid command stream. RESP Array index %d exceeds specified capacity of %s", counter+1, numberOfItems.ToString()))
		}
		first := bytes[0]
		switch first {
		case stringStartByte:
			s, r = parseSimpleString(bytes)
		case integerStartByte:
			s, r = parseIntegers(bytes)
		case bulkStringStartByte:
			s, r = parseBulkString(bytes)
		default:
			panic("Unknown start byte " + string(first))
		}
		// Append to chunks
		Array.SetItemAtIndex(counter, s)
		bytes = bytes[r:]
		bytesRead += r
		counter++
	}
	return Array, bytesRead
}

// Get the next array start byte starting from offset 1
func getNextArrayStartByteIndex(bytes []byte) int {
	for i := 1; i < len(bytes); i++ {
		if bytes[i] == arrayStartByte {
			return i
		}
	}
	return len(bytes)
}

// ParseRedisClientRequest takes in a sequence of bytes, and parses them
// as sequential Array entries. Each command in a pipeline will form
// a Array. This method catches internal panics, and returns top level
// errors as RedisError. The caller can then check if the error is EmptyRedisError
// and return appropriately
func ParseRedisClientRequest(bytes []byte) ([]Array, int, RedisError) {
	commands := make([]Array, 0)
	totalBytesRead := 0
	finalErr := EmptyRedisError

	// Top level panic recovery
	defer func() {
		if r := recover(); r != nil {
			switch re := r.(type) {
			case RedisError:
				finalErr = re
			case string:
				finalErr = NewRedisError(DefaultErrorKeyword, re)
			default:
				fmt.Println(r)
				finalErr = NewDefaultRedisError(fmt.Sprint(r))
			}
		}
	}()
	//Continue to read until all input request parsed
	for len(bytes) > 0 {
		//Assumption,Clients send commands to the Redis server using RESP Arrays
		asbIndex := getNextArrayStartByteIndex(bytes)

		command, read := parseArray(bytes[0:asbIndex])
		if read > 0 {
			commands = append(commands, *command)
			bytes = bytes[read:]
			totalBytesRead += read
		} else {
			break
		}
	}
	return commands, totalBytesRead, finalErr
}
