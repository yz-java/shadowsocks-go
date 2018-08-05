/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 14:47
 */
package model

import "time"

type ClientConfig struct {

	Id int

	CPort int

	CPassword string

	CStatus int

	CreateTime time.Time
}

