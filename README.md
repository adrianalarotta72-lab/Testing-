package store

import (
	"errors"
)

// Errors
var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrEmptyKey         = errors.New("key cannot be empty")
	ErrKeyAlreadyExists = errors.New("key already exists")
)

// StoreOperation representa una operación a realizar
type StoreOperation struct {
	Type     string // "SET", "UPDATE", "GET", "DELETE", "EXISTS", "SIZE", "CLEAR"
	Key      string
	Value    string
	RespChan chan StoreResponse
}

// StoreResponse representa la respuesta de una operación
type StoreResponse struct {
	Value  string
	Exists bool
	Size   int
	Error  error
}

// Canal global para operaciones del store
var storeChan chan StoreOperation

// init inicializa el canal y arranca el goroutine del store
func init() {
	storeChan = make(chan StoreOperation)
	go run()
}

// run es el goroutine que procesa todas las operaciones
// Solo este goroutine accede al map, por lo tanto no necesita mutex
func run() {
	data := make(map[string]string)

	for op := range storeChan {
		switch op.Type {
		case "SET":
			handleSet(data, op)
		case "UPDATE":
			handleUpdate(data, op)
		case "GET":
			handleGet(data, op)
		case "DELETE":
			handleDelete(data, op)
		case "EXISTS":
			handleExists(data, op)
		case "SIZE":
			handleSize(data, op)
		case "CLEAR":
			handleClear(data, op)
		}
	}
}

func handleSet(data map[string]string, op StoreOperation) {
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	// Verifica si la key ya existe
	if _, exists := data[op.Key]; exists {
		op.RespChan <- StoreResponse{Error: ErrKeyAlreadyExists}
		return
	}

	data[op.Key] = op.Value
	op.RespChan <- StoreResponse{Error: nil}
}

func handleUpdate(data map[string]string, op StoreOperation) {
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	// Verifica si la key existe
	if _, exists := data[op.Key]; !exists {
		op.RespChan <- StoreResponse{Error: ErrKeyNotFound}
		return
	}

	data[op.Key] = op.Value
	op.RespChan <- StoreResponse{Error: nil}
}

func handleGet(data map[string]string, op StoreOperation) {
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	value, exists := data[op.Key]
	if !exists {
		op.RespChan <- StoreResponse{Error: ErrKeyNotFound}
		return
	}

	op.RespChan <- StoreResponse{Value: value, Error: nil}
}

func handleDelete(data map[string]string, op StoreOperation) {
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	if _, exists := data[op.Key]; !exists {
		op.RespChan <- StoreResponse{Error: ErrKeyNotFound}
		return
	}

	delete(data, op.Key)
	op.RespChan <- StoreResponse{Error: nil}
}

func handleExists(data map[string]string, op StoreOperation) {
	_, exists := data[op.Key]
	op.RespChan <- StoreResponse{Exists: exists, Error: nil}
}

func handleSize(data map[string]string, op StoreOperation) {
	op.RespChan <- StoreResponse{Size: len(data), Error: nil}
}

func handleClear(data map[string]string, op StoreOperation) {
	for k := range data {
		delete(data, k)
	}
	op.RespChan <- StoreResponse{Error: nil}
}

// === FUNCIONES PÚBLICAS A NIVEL DE PAQUETE ===

// Set guarda un key-value (NO sobrescribe si ya existe)
func Set(key, value string) error {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "SET",
		Key:      key,
		Value:    value,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Error
}

// Update actualiza un key-value existente (SÍ sobrescribe)
func Update(key, value string) error {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "UPDATE",
		Key:      key,
		Value:    value,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Error
}

// Get obtiene el valor de una key
func Get(key string) (string, error) {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "GET",
		Key:      key,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Value, resp.Error
}

// Delete elimina una key
func Delete(key string) error {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "DELETE",
		Key:      key,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Error
}

// Exists verifica si una key existe
func Exists(key string) bool {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "EXISTS",
		Key:      key,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Exists
}

// Size retorna el número de elementos
func Size() int {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "SIZE",
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Size
}

// Clear elimina todos los elementos
func Clear() {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "CLEAR",
		RespChan: respChan,
	}
	<-respChan
}
