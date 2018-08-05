/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-05
 * Time: 14:42
 */
package db

import "testing"

func TestCreateEngine(t *testing.T) {
	CreateEngine()
	StartAutoConnect()
}
