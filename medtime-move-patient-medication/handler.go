package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
)

const (
	apiKey = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"

	getListURL   = "https://api.admin.u-code.io/v1/object-slim/get-list/"
	getSingleURL = "https://api.admin.u-code.io/v1/object/"
	// multipleUpdateUrl    = "https://api.admin.u-code.io/v1/object/multiple-update/"
	// getListObjectBuilder = "https://api.admin.u-code.io/v1/object/get-list/"
)

// ! SERVER NOW TIME = TASHKENT TIME - 5

// func main() {
// 	Handle([]byte{})
// }

// Handle a serverless request
func Handle(req []byte) string {

	err := CreateNewMedication()
	if err != nil {
		return Handler("error", "CreateNewMedication >>>> "+err.Error())
	}

	err = UpdateTakeTime()
	if err != nil {
		return Handler("error", "UpdateTakeTime >>>>>> "+err.Error())
	}

	return Handler("OK", "Success")
}

func CreateNewMedication() error {

	t := time.Now()

	Handler("info", "CreateNewMedication >>>> ")

	var (
		getMedicineUrl  = getListURL + "medicine_taking" + `?data={"frequency":["always"],"is_from_patient":true}`
		getMedicineResp = GetListClientApiResponse{}
	)

	body, err := DoRequest(getMedicineUrl, "GET", nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &getMedicineResp); err != nil {
		return err
	}

	var wg sync.WaitGroup

	Handler("info", "CreateNewMedication len >>>>> "+fmt.Sprint(len(getMedicineResp.Data.Data.Response)))

	for _, medicine := range getMedicineResp.Data.Data.Response {

		wg.Add(1)

		go func(medicine map[string]interface{}) {
			defer wg.Done()
			var (
				deleteMedicineUrl = getSingleURL + "medicine_taking/" + fmt.Sprint(medicine["guid"])
			)

			_, err := DoRequest(deleteMedicineUrl, "DELETE", Request{Data: map[string]interface{}{}})
			if err != nil {
				return
			}

			delete(medicine, "guid")

			var (
				createMedicineUrl = getSingleURL + "medicine_taking"
				createMedicineReq = Request{
					Data: medicine,
				}
			)

			_, err = DoRequest(createMedicineUrl, "POST", createMedicineReq)
			if err != nil {
				return
			}

		}(medicine)
	}

	wg.Wait()

	Handler("info", "CreateNewMedication >>>> "+fmt.Sprint(time.Since(t).String()))

	return nil
}

func UpdateTakeTime() error {

	t := time.Now()

	Handler("info", "UpdateTakeTime >>>> ")

	var (
		getMedicineTUrl  = getListURL + "medicine_taking" + `?data={"frequency":["specific_days","several_times_day"]}`
		getMedicineTResp = GetListClientApiResponse{}
	)

	body, err := DoRequest(getMedicineTUrl, "GET", nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &getMedicineTResp); err != nil {
		return err
	}

	Handler("info", "UpdateTakeTime len >>>> "+fmt.Sprint(len(getMedicineTResp.Data.Data.Response)))

	var wg sync.WaitGroup

	for _, medicine := range getMedicineTResp.Data.Data.Response {

		wg.Add(1)

		go func(medicine map[string]interface{}) {
			defer wg.Done()

			var (
				getPatientUrl  = getListURL + "patient_medication" + fmt.Sprintf(`?data={"medicine_taking_id":"%s","is_take":true}`, medicine["guid"])
				getPatientResp = GetListClientApiResponse{}
				medicineId     = uuid.NewString()
				currentAmount  = medicine["current_amount"]
			)

			body, err := DoRequest(getPatientUrl, "GET", nil)
			if err != nil {
				return
			}
			if err := json.Unmarshal(body, &getPatientResp); err != nil {
				return
			}

			var (
				deleteMedicineUrl = getSingleURL + "medicine_taking/" + fmt.Sprint(medicine["guid"])
			)

			_, err = DoRequest(deleteMedicineUrl, "DELETE", Request{Data: map[string]interface{}{}})
			if err != nil {
				return
			}

			medicine["current_amount"] = medicine["current_left"]
			medicine["guid"] = medicineId

			delete(medicine, "current_left")

			var (
				createMedicineUrl = getSingleURL + "medicine_taking"
				createMedicineReq = Request{
					Data: medicine,
				}
			)

			_, err = DoRequest(createMedicineUrl, "POST", createMedicineReq)
			if err != nil {
				return
			}

			for _, patient := range getPatientResp.Data.Data.Response {

				patient["medicine_taking_id"] = medicineId

				var (
					createPatientUrl = getSingleURL + "patient_medication"
					createPatientReq = Request{
						Data: patient,
					}
				)

				_, err = DoRequest(createPatientUrl, "POST", createPatientReq)
				if err != nil {
					return
				}
			}

			createMedicineReq.Data["current_amount"] = currentAmount

			_, err = DoRequest(createMedicineUrl, "PUT", createMedicineReq)
			if err != nil {
				return
			}

		}(medicine)

		wg.Wait()

	}

	Handler("info", "UpdateTakeTime >>>>> "+fmt.Sprint(time.Since(t).String()))

	return nil
}
