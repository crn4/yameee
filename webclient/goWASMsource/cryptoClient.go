package main

import (
	"mytests/yetanothermessenger/server/cryprot"
	"syscall/js"
)

func main() {
	js.Global().Set("generateKeys", js.FuncOf(GeneratePubPrivKeys))
	js.Global().Set("calculateSecret", js.FuncOf(calculateSecretKey))
	js.Global().Set("encryptMess", js.FuncOf(encryptMessage))
	js.Global().Set("decryptMess", js.FuncOf(decryptMessage))
	<-make(chan bool)
}

func encryptMessage(this js.Value, inputs []js.Value) interface{} {
	messageString := inputs[0].String()
	message := []byte(messageString)
	secretKey := make([]byte, inputs[1].Get("length").Int())
	js.CopyBytesToGo(secretKey, inputs[1])
	cipherMessage := cryprot.EncryptMessage(message, secretKey)
	JScipherMessage := js.Global().Get("Uint8Array").New(len(cipherMessage))
	js.CopyBytesToJS(JScipherMessage, cipherMessage)
	return JScipherMessage
}

func decryptMessage(this js.Value, inputs []js.Value) interface{} {
	cipherMessage := make([]byte, inputs[0].Get("length").Int())
	js.CopyBytesToGo(cipherMessage, inputs[0])
	secretKey := make([]byte, inputs[1].Get("length").Int())
	js.CopyBytesToGo(secretKey, inputs[1])
	message := cryprot.DecryptMessage(cipherMessage, secretKey)
	return string(message)
}

func GeneratePubPrivKeys(this js.Value, inputs []js.Value) interface{} {
	publicKey, privateKey := cryprot.GeneratePubPrivKeys()

	JSpublicKey := js.Global().Get("Uint8Array").New(len(publicKey))
	js.CopyBytesToJS(JSpublicKey, publicKey)
	JSprivateKey := js.Global().Get("Uint8Array").New(len(privateKey))
	js.CopyBytesToJS(JSprivateKey, privateKey)

	return map[string]interface{}{
		"publicKey":  JSpublicKey,
		"privateKey": JSprivateKey,
	}
}

func calculateSecretKey(this js.Value, inputs []js.Value) interface{} {
	foreignKey := make([]byte, inputs[0].Get("length").Int())
	privateKey := make([]byte, inputs[1].Get("length").Int())
	js.CopyBytesToGo(foreignKey, inputs[0])
	js.CopyBytesToGo(privateKey, inputs[1])

	secretKey := cryprot.CalcSecretKey(foreignKey, privateKey)
	JSsecretKey := js.Global().Get("Uint8Array").New(len(secretKey))
	js.CopyBytesToJS(JSsecretKey, secretKey)

	return JSsecretKey
}
