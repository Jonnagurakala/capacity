// Code generated by go-swagger; DO NOT EDIT.

package workers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/supergiant/capacity/pkg/capacityclient/models"
)

// ListWorkersReader is a Reader for the ListWorkers structure.
type ListWorkersReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListWorkersReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewListWorkersOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewListWorkersOK creates a ListWorkersOK with default headers values
func NewListWorkersOK() *ListWorkersOK {
	return &ListWorkersOK{}
}

/*ListWorkersOK handles this case with default header values.

workerListResponse contains a list of workers.
*/
type ListWorkersOK struct {
	Payload *models.WorkerList
}

func (o *ListWorkersOK) Error() string {
	return fmt.Sprintf("[GET /api/v1/workers][%d] listWorkersOK  %+v", 200, o.Payload)
}

func (o *ListWorkersOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.WorkerList)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
