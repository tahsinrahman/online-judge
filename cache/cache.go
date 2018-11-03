package cache

import (
	"encoding/json"
	"reflect"

	"github.com/go-macaron/cache"
	"github.com/go-xorm/xorm"
	"github.com/tahsinrahman/online-judge/db"
)

func StoreCache(c cache.Cache, key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// store in cache
	c.Put(key, string(b), 3600*24)
	return nil
}

func CheckCache(c cache.Cache, key string, value interface{}) (bool, error) {
	cache := c.Get(key)
	if cache != nil {
		if err := json.Unmarshal([]byte(cache.(string)), &value); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// find user by username
func FindObject(c cache.Cache, key string, value interface{}, dbSession *xorm.Session, get bool) (bool, error) {
	// first check in cache
	has, err := CheckCache(c, key, value)
	if err != nil {
		return false, err
	}
	if has {
		return true, nil
	}

	// not found in cache
	// search in database
	if get {
		if has, err = db.Engine.Get(value); err != nil {
			return false, err
		} else if !has {
			return false, nil
		}
	} else {
		if err = dbSession.Find(value); err != nil {
			return false, err
		} else if value == nil {
			return false, nil
		}
	}

	if err = StoreCache(c, key, value); err != nil {
		return false, err
	}

	return true, nil
}

func AddToList(key string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if err = db.Client.SAdd(key, b).Err(); err != nil {
		return err
	}
	return nil
}

func CheckList(key string, v interface{}, dbSession *xorm.Session) error {
	if db.Client.Exists(key).Val() == int64(0) {
		// get list from db
		if err := dbSession.Find(v); err != nil {
			return err
		}

		// store it in cache
		slice := reflect.Indirect(reflect.ValueOf(v))
		len := slice.Len()
		for i := 0; i < len; i++ {
			if err := AddToList(key, slice.Index(i).Interface()); err != nil {
				return err
			}
		}
	} else {
		// get list from cache
		results, err := db.Client.SMembers(key).Result()
		if err != nil {
			return err
		}

		slice := reflect.Indirect(reflect.ValueOf(v))

		// unmarshall it and save it to v
		for _, str := range results {
			val := reflect.New(slice.Type().Elem())
			if err = json.Unmarshal([]byte(str), val.Interface()); err != nil {
				return err
			}
			slice.Set(reflect.Append(slice, reflect.Indirect(val)))
		}
	}
	return nil
}

// user_USERNAME
// contest_ID
// contest_ID_problem_PID
// problem_ID_dataset
// submissions_user_USERNAME_problem_ID
// perm_contest_ID_user_USERNAME
// submission_ID
// contest_list
// contest_ID_problems
// contest_ID_submissions
