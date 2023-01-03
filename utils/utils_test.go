package utils

// unit test for utils package
// func name must be Testxxx (xxx is )
import (
	"database/sql/driver"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestCount(t *testing.T) {
	tables := []struct {
		a []int
		n int
	}{
		{[]int{1, 1}, 2},
		{[]int{1, 2}, 3},
		{[]int{2, 2}, 4},
		{[]int{5, 2}, 7},
	}

	for _, table := range tables {
		total := Sum(table.a)
		if total != table.n {
			t.Errorf("Sum of (%d+%d) was incorrect, got: %d, want: %d.", table.a[0], table.a[1], total, table.n)
		}
	}
}

type CustomerT struct {
	Name string `json:"name"`
	Dob  string `json:"dob"`
	Note string `json:"note"`
}

type ListCustomerT []CustomerT

func (item *ListCustomerT) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListCustomerT) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func TestReadData(t *testing.T) {
	// Get public key from local
	f, errF := os.Open("../customer.json")
	if errF != nil {
		log.Println("TestReadData errF", errF.Error())
		return
	}
	defer f.Close()

	byteCustomerValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("TestReadData errRA", errRA.Error())
		return
	}

	listData := ListCustomerT{}

	errUnM := json.Unmarshal(byteCustomerValue, &listData)
	if errUnM != nil {
		log.Println("TestReadData errUnM", errUnM.Error())
		return
	}
	log.Println("ok")
	log.Print(listData)

}
