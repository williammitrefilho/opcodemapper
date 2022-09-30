module modgo.com/modgo

go 1.19

replace modgo.com/isok => /isok

require (
	modgo.com/server v0.0.0-00010101000000-000000000000
	modgo.com/whatserver v0.0.0-00010101000000-000000000000
)

require modgo.com/what v0.0.0-00010101000000-000000000000 // indirect

replace modgo.com/what => /what

replace modgo.com/server => /server

replace modgo.com/whatserver => /whatserver
