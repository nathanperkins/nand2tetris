// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Start by making the screen white
	@WHITE
	0; JMP

(TOP)
// Write value to all positions

// for (i = 0; i < 32; i++)

// i = 0
	@i
	M=0

// addr = &SCREEN
	@SCREEN
	D=A
	@addr
	M=D 

(FILL)
// i < 32
	@i
	D=M
	@8192 // number of words in the screen
	D=D-A
	@CHECK
	D; JGE

// Mem[addr] = value
	@value
	D=M
	@addr
	A=M
	M=D

// i++
	@i
	M=M+1
// addr++
	@addr
	M=M+1
	@FILL
	0; JMP

// check keyboard register
(CHECK)
	@KBD
	D=M
	@BLACK
	D; JNE

// white, if button is pressed
(WHITE)
	@value
	M=0
	@TOP
	0; JMP

// black, otherwise
(BLACK)
	@value
	M=-1
	@TOP
	0; JMP
