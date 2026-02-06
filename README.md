package store

import (
	"encoding/json"
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
	Type     string
	Key      string
	Value    interface{} // Soporta cualquier tipo de dato
	RespChan chan StoreResponse
}

// StoreResponse representa la respuesta de una operación
type StoreResponse struct {
	Value  interface{} // Puede retornar cualquier tipo
	Exists bool
	Size   int
	Error  error
}

// Canal global para las operaciones del store
var storeChan chan StoreOperation

// Canal para shutdown graceful
var shutdownChan chan struct{}

// init inicializa el store automáticamente cuando se importa el paquete
func init() {
	storeChan = make(chan StoreOperation)
	shutdownChan = make(chan struct{})
	go runStore()
}

// runStore es el único goroutine que accede al map
// Elimina completamente la necesidad de mutex
func runStore() {
	data := make(map[string]interface{}) // Soporta cualquier tipo

	for {
		select {
		case op := <-storeChan:
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
		case <-shutdownChan:
			// Graceful shutdown
			return
		}
	}
}

// === HANDLERS INTERNOS ===

func handleSet(data map[string]interface{}, op StoreOperation) {
	// Validación temprana (code review feedback)
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	// NO sobrescribe si existe
	if _, exists := data[op.Key]; exists {
		op.RespChan <- StoreResponse{Error: ErrKeyAlreadyExists}
		return
	}

	data[op.Key] = op.Value
	op.RespChan <- StoreResponse{Error: nil}
}

func handleUpdate(data map[string]interface{}, op StoreOperation) {
	// Validación temprana
	if op.Key == "" {
		op.RespChan <- StoreResponse{Error: ErrEmptyKey}
		return
	}

	// SÍ sobrescribe, pero debe existir
	if _, exists := data[op.Key]; !exists {
		op.RespChan <- StoreResponse{Error: ErrKeyNotFound}
		return
	}

	data[op.Key] = op.Value
	op.RespChan <- StoreResponse{Error: nil}
}

func handleGet(data map[string]interface{}, op StoreOperation) {
	// Validación temprana
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

func handleDelete(data map[string]interface{}, op StoreOperation) {
	// Validación temprana
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

func handleExists(data map[string]interface{}, op StoreOperation) {
	_, exists := data[op.Key]
	op.RespChan <- StoreResponse{Exists: exists, Error: nil}
}

func handleSize(data map[string]interface{}, op StoreOperation) {
	op.RespChan <- StoreResponse{Size: len(data), Error: nil}
}

func handleClear(data map[string]interface{}, op StoreOperation) {
	for k := range data {
		delete(data, k)
	}
	op.RespChan <- StoreResponse{Error: nil}
}

// === FUNCIONES PÚBLICAS A NIVEL DE PAQUETE ===

// Set guarda un key-value (NO sobrescribe si ya existe)
// Acepta cualquier tipo de dato como value (string, int, bool, map, slice, etc.)
func Set(key string, value interface{}) error {
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
// Acepta cualquier tipo de dato como value
func Update(key string, value interface{}) error {
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
// Retorna interface{}, el llamador debe hacer type assertion
func Get(key string) (interface{}, error) {
	respChan := make(chan StoreResponse)
	storeChan <- StoreOperation{
		Type:     "GET",
		Key:      key,
		RespChan: respChan,
	}
	resp := <-respChan
	return resp.Value, resp.Error
}

// GetAsString obtiene el valor y lo convierte a string
// Si el valor no es string, lo serializa como JSON
func GetAsString(key string) (string, error) {
	value, err := Get(key)
	if err != nil {
		return "", err
	}

	// Si ya es string, retornarlo directamente
	if str, ok := value.(string); ok {
		return str, nil
	}

	// Si es otro tipo, convertir a JSON
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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

// Shutdown cierra gracefully el store
func Shutdown() {
	close(shutdownChan)
}
