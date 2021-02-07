package commands

// Uses a subset of commands from https://redis.io/commands#string + DEL command

import (
	"fmt"
	"golang-redis/resp"
	"golang-redis/storage"
)

const (
	getCommand    = "GET"
	setCommand    = "SET"
	deleteCommand = "DEL"
)

var gMap = storage.NewGenericConcurrentMap()
var redisOk = resp.NewString("OK")

// execute a get command on concurrent map and return the result
func executeGetCommand(ra *resp.Array) (resp.IDataType, resp.RedisError) {
	// 	// Get argument takes only a single key name.
	numberOfItems := ra.GetNumberOfItems()
	if numberOfItems == 1 {
		return nil, resp.NewDefaultRedisError("wrong number of arguments for (get) command")
	} else if numberOfItems > 2 {
		// First item is the command itself
		// Ignore with warning message
		fmt.Printf("WARN: GET command acccepts only one argument. But received %d. Other arguments will be ignored\n", numberOfItems-1)
	}
	key, err := getGuardedKey(ra.GetItemAtIndex(1))
	if err != resp.EmptyRedisError {
		return nil, err
	}
	value, ok := gMap.Get(key)
	if ok != true {
		// If we cannot find it, we return Nil bulk string
		return resp.EmptyBulkString, resp.EmptyRedisError
	}
	bs, e := resp.NewBulkString(value)
	if e != nil {
		return nil, resp.NewDefaultRedisError(e.Error())
	}
	return bs, resp.EmptyRedisError
}

//Reply to client with valid response of its set request
func setCommandReply(v resp.IDataType, returnPreviousKey bool) (resp.IDataType, resp.RedisError) {
	if returnPreviousKey == true {
		// Fetch previous key value
		bs, e := resp.NewBulkString(v.ToString())
		if e != nil {
			return resp.EmptyBulkString, resp.NewDefaultRedisError(e.Error())
		}
		return bs, resp.EmptyRedisError
	}
	// Otherwise return 'OK' as bulk string
	return redisOk, resp.EmptyRedisError
}

// Guarded key check to verify that key is string
func getGuardedKey(key resp.IDataType) (string, resp.RedisError) {
	switch key.(type) {
	case resp.String:
		return key.ToString(), resp.EmptyRedisError
	case resp.BulkString:
		return key.ToString(), resp.EmptyRedisError
	default:
		return "", resp.NewDefaultRedisError(fmt.Sprintf("%s expects a string key value", getCommand))
	}
}

// execute a set command on concurrent map. If returnPreviousKey is set to true, then it returns
// the previous set value as first return value
func executeSetCommand(ra *resp.Array, returnPreviousKey bool, onlyIfKeyExists bool) (resp.IDataType, resp.RedisError) {
	numberOfItems := ra.GetNumberOfItems()
	if numberOfItems == 2 {
		return resp.EmptyBulkString, resp.NewDefaultRedisError("wrong number of arguments for (set) command")
	} else if numberOfItems > 3 {
		// First item is the command itself
		// Second is key
		// Last is value
		// Ignore with warning message
		fmt.Printf("WARN: SET command acccepts only two arguments. But received %d\n. Other arguments will be ignored", numberOfItems-1)
	}
	key, err := getGuardedKey(ra.GetItemAtIndex(1))
	if err != resp.EmptyRedisError {
		return resp.EmptyBulkString, err
	}
	value := ra.GetItemAtIndex(2)
	if onlyIfKeyExists {
		_, ok := gMap.Get(key)
		if ok != true {
			gMap.Set(key, value.ToString())
			return resp.NewInteger(1), resp.EmptyRedisError
		} else {
			return resp.NewInteger(0), resp.EmptyRedisError
		}
	}
	gMap.Set(key, value.ToString())
	return setCommandReply(value, returnPreviousKey)
}

// Delete a key from storage, and return number of keys removed
func executeDeleteCommand(ra *resp.Array) (resp.Integer, resp.RedisError) {
	// Get number of items
	numberOfItems := ra.GetNumberOfItems()
	numberOfKeysDeleted := 0
	if numberOfItems == 1 {
		return resp.EmptyInteger, resp.NewDefaultRedisError("wrong number of arguments for (del) command")
	}
	for k := 1; k < numberOfItems; k++ {
		key, err := getGuardedKey(ra.GetItemAtIndex(1))
		if err != resp.EmptyRedisError {
			return resp.EmptyInteger, resp.NewDefaultRedisError(fmt.Sprintf("%s expects a string key value", getCommand))
		}
		ok := gMap.Delete(key)
		if ok == true {
			numberOfKeysDeleted++
		}
	}
	return resp.NewInteger(numberOfKeysDeleted), resp.EmptyRedisError
}

// ExecuteStringCommand takes a Array and inspects it to check there is
// a matching executable command. If no command can be found, it returns error
func ExecuteStringCommand(ra resp.Array) (resp.IDataType, resp.RedisError) {
	if ra.GetNumberOfItems() == 0 {
		return nil, resp.NewDefaultRedisError("No command found")
	}
	first := ra.GetItemAtIndex(0)
	switch first.ToString() {
	case getCommand:
		return executeGetCommand(&ra)
	case setCommand:
		return executeSetCommand(&ra, false, false)
	case deleteCommand:
		return executeDeleteCommand(&ra)
	default:
		break
	}
	return nil, resp.NewDefaultRedisError(fmt.Sprintf("Unknown or disabled command '%s'", first.ToString()))
}
