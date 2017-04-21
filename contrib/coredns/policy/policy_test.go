package policy

import (
	"testing"
	"errors"
	"github.com/miekg/dns"
	"net"

)



// This function tests that Validation of Middleware response is working properly
// Inputs are t *testing.T
// Output is the error update in t
func TestValidateMiddlewareResponse(t *testing.T) {
	var p PolicyMiddleware
	var dnsResponseWriter	dns.ResponseWriter

	dnsMsg := new(dns.Msg)
	//
	// Case 1
	//
	var response1 Response

	// Case 1a - Err is set, return should have err and RcodeServerFailure
	err := errors.New("Invalid error")
	status := dns.RcodeServerFailure
	returnedStatus, returnedError := ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg, &response1, status, err)
	if returnedError == nil || returnedStatus != dns.RcodeServerFailure {
		t.Fatalf("Expected error != nil and response RcodeServerFailure but got %#v,%#v", returnedError,returnedStatus)
	}

	// Case 1b - Err is set to invalid and status is succes
	err = errors.New("Invalid error")
	status = dns.RcodeSuccess
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg, &response1, status, err)
	if returnedError == nil || returnedStatus != dns.RcodeServerFailure {
		t.Fatalf("Expected error != nil and response RcodeServerFailure but got %#v,%#v", returnedError,returnedStatus)
	}

	//
	// Case 2
	//
	var response2 Response

	// Case 2a - Permit is set to true and Redirect to nil
	response2.Permit = true
	response2.Redirect = nil
	status = dns.RcodeSuccess
	err = nil
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg,&response2, status, err)
	if returnedError != nil && returnedStatus != status {
		t.Fatalf("Expected error nil and response RcodeServerFailure but got %#v,%#v", returnedError,returnedStatus)
	}

	// Permit is true, Redirect is nil, status is Success, err is nil
	response2.Permit = true
	response2.Redirect = nil
	status = dns.RcodeRefused
	err = nil
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg, &response2, status, err)
	if returnedError != nil && returnedStatus != status {
		t.Fatalf("Expected error nil and response RcodeRefused but got %#v,%#v", returnedError,returnedStatus)
	}

	//
	// Case 3
	//
	var response3 Response
	dnsMsg.SetQuestion("example.com.", dns.TypeA)
	dnsMsg.SetEdns0(4097, true)
	// Case 3a : Just set Redirect to non null
	response3.Redirect = net.IP("10.10.10.1")
	err = nil
	status = dns.RcodeSuccess
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg, &response3, status, err)
	if returnedError != nil {
		t.Fatalf("Expected error nil but got %#v", returnedError)
	}

	//
	// Case 4
	//
	var response4 Response
	// Case 4a
	response4.Permit = false
	status = dns.RcodeRefused
	err = nil
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, dnsMsg, &response4, status, err)
	if returnedError == nil || returnedStatus != dns.RcodeRefused {
		t.Fatalf("Expected error != nil and response = RcodeRefused but got %#v,%#v", returnedError,returnedStatus)
	}
	// Case 4b
	response4.Permit = false
	response4.Redirect = nil
	status = dns.RcodeSuccess
	err = nil
	returnedStatus, returnedError = ValidateMiddlewareResponse(&p, dnsResponseWriter, &dnsMsg, &response4, status, err)
	if returnedError == nil || returnedStatus != dns.RcodeRefused {
		t.Fatalf("Expected error != nil and response = RcodeRefused but got %#v,%#v", returnedError,returnedStatus)
	}

}

