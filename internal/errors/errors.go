package errors

import "log"

func PanicOnError(err error, moreInfo *string) {
	if err != nil {
		if moreInfo != nil {
			log.Println(*moreInfo)
		}
		panic(err)
	}
}
