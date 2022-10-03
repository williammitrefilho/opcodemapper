# An x86-64 assembler/disassembler

For disassembly:

**POST** request body:
```json
{
	"code":"6858AC4000E8EEFFFFFF"
}
```

response:
```json
{
	"code":[
		"PUSH 0x0040AC58",
		"CALL [RIP] + 0xFFFFFFEE"
	]
}
```

In progress.
Being implemented as defined in [The Manual](https://www.amd.com/system/files/TechDocs/24594.pdf)