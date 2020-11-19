package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

type API_Server struct {
	started     int32
	shutdown    int32
	statusLock  sync.RWMutex
	statusLines map[int]string
}

var timeZeroVal time.Time

type JsonRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

func parseJsonRequest(rawMessage []byte, method string) (*JsonRequest, error) {
	var request JsonRequest
	if len(rawMessage) == 0 && method == "POST" {
		fmt.Println("Method - " + method)
		return &request, rpcservice.NewRPCError(rpcservice.RPCParseError, nil)
	}
	err := json.Unmarshal(rawMessage, &request)
	if err != nil {
		// Logger.log.Error("Can not parse", string(rawMessage))
		return &request, rpcservice.NewRPCError(rpcservice.RPCParseError, err)
	} else {
		return &request, nil
	}
}

func (api *API_Server) Start() {
	api.statusLines = make(map[int]string)
}

func (api *API_Server) ProcessRpcRequest(w http.ResponseWriter, r *http.Request, isLimitedUser bool) {
	if atomic.LoadInt32(&api.shutdown) != 0 {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON Message: %+v", errCode, err), errCode)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		// Logger.log.Error(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		// Logger.log.Errorf("Failed to hijack HTTP connection: %s", err.Error())
		// Logger.log.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()
	conn.SetReadDeadline(timeZeroVal)

	var jsonErr error
	var result interface{}
	var request *JsonRequest
	request, jsonErr = parseJsonRequest(body, r.Method)

	if jsonErr == nil {
		// if request.Id == nil && !(httpServer.config.RPCQuirks && request.Jsonrpc == "") {
		// 	return
		// }

		// if httpServer.config.RPCLimitRequestErrorPerHour > 0 {
		// 	if httpServer.checkBlackListClientRequestErrorPerHour(r, request.Method) {
		// 		errMsg := "Reach limit request error for method " + request.Method
		// 		Logger.log.Error(errMsg)
		// 		errCode := http.StatusTooManyRequests
		// 		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		// 		return
		// 	}
		// }

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()

		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			command := HttpHandler[request.Method]
			if command == nil {
				result = nil
				jsonErr = rpcservice.NewRPCError(rpcservice.RPCMethodNotFoundError, errors.New("Method not found: "+request.Method))
			}
			if command != nil {
				result, jsonErr = command(api, request.Params, closeChan)
			} else {
				jsonErr = rpcservice.NewRPCError(rpcservice.RPCMethodNotFoundError, errors.New("Method not found: "+request.Method))
			}
		}
	}

	if jsonErr.(*rpcservice.RPCError) != nil && r.Method != "OPTIONS" {
		if jsonErr.(*rpcservice.RPCError).Code == rpcservice.ErrCodeMessage[rpcservice.RPCParseError].Code {
			// Logger.log.Errorf("RPC function process with err \n %+v", jsonErr)
			api.writeHTTPResponseHeaders(r, w.Header(), http.StatusBadRequest, buf)
			// httpServer.addBlackListClientRequestErrorPerHour(r, request.Method)
			return
		}
	}

	// if jsonErr != nil && jsonErr.(*rpcservice.RPCError) != nil {
	// 	httpServer.addBlackListClientRequestErrorPerHour(r, request.Method)
	// }

	// Marshal the response.
	msg, err := createMarshalledResponse(request, result, jsonErr)
	if err != nil {
		// Logger.log.Errorf("Failed to marshal reply: %s", err.Error())
		// Logger.log.Error(err)
		return
	}

	// Write the response.
	// for testing only
	// w.WriteHeader(http.StatusOK)
	err = api.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		// Logger.log.Error(err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		// Logger.log.Errorf("Failed to write marshalled reply: %s", err.Error())
		// Logger.log.Error(err)
	}

	// Terminate with newline to maintain compatibility with coin Core.
	if err := buf.WriteByte('\n'); err != nil {
		// Logger.log.Errorf("Failed to append terminating newline to reply: %s", err.Error())
		// Logger.log.Error(err)
	}
}
func (api *API_Server) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	api.statusLock.RLock()
	line, ok := api.statusLines[key]
	api.statusLock.RUnlock()
	if ok {
		return line
	}

	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		api.statusLock.Lock()
		api.statusLines[key] = line
		api.statusLock.Unlock()
	} else {
		text = "status Code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}
	return line
}

func (api *API_Server) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, api.httpStatusLine(req, code))
	if err != nil {
		return err
	}
	err = headers.Write(w)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\r\n")
	return err
}

func createMarshalledResponse(request *JsonRequest, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *rpcservice.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*rpcservice.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = rpcservice.InternalRPCError(replyErr.Error(), "")
		}
	}
	// MarshalResponse marshals the passed id, result, and RPCError to a JSON-RPC
	// response byte slice that is suitable for transmission to a JSON-RPC client.
	marshalledResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	response, err := newResponse(request, marshalledResult, jsonErr)
	if err != nil {
		return nil, err
	}
	resultResp, err := json.MarshalIndent(&response, "", "\t")
	if err != nil {
		return nil, err
	}
	return resultResp, nil
}

type JsonResponse struct {
	Id      *interface{}         `json:"Id"`
	Result  json.RawMessage      `json:"Result"`
	Error   *rpcservice.RPCError `json:"Error"`
	Params  interface{}          `json:"Params"`
	Method  string               `json:"Method"`
	Jsonrpc string               `json:"Jsonrpc"`
}

func newResponse(request *JsonRequest, marshalledResult []byte, rpcErr *rpcservice.RPCError) (*JsonResponse, error) {
	id := request.Id
	if !IsValidIDType(id) {
		str := fmt.Sprintf("The id of type '%T' is invalid", id)
		return nil, rpcservice.NewRPCError(rpcservice.InvalidTypeError, errors.New(str))
	}
	pid := &id
	resp := &JsonResponse{
		Id:      pid,
		Result:  marshalledResult,
		Error:   rpcErr,
		Params:  request.Params,
		Method:  request.Method,
		Jsonrpc: request.Jsonrpc,
	}
	if resp.Error != nil {
		resp.Error.StackTrace = rpcErr.Error()
	}
	return resp, nil
}

func NewCorsHeader(w http.ResponseWriter) {
	// Set CORS Header
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Device-Type, Device-Id, Authorization, Accept-Language, Access-Control-Allow-Headers, Access-Control-Allow-Credentials, Access-Control-Allow-Origin, Access-Control-Allow-Methods, *")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, GET, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
