package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func openFile(filePath string) string {
	buffer, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file `%s`: %s\n", filePath, err)
		os.Exit(1)
	}
	buffer = append(buffer, '\000')

	return string(buffer)
}

func shaderTypeAsStr(shader uint32) string {
	switch shader {
	case gl.VERTEX_SHADER:
		return "GL_VERTEX_SHADER"
	case gl.FRAGMENT_SHADER:
		return "GL_FRAGMENT_SHADER"
	default:
		return "(Unknown)"
	}
}

func compileShaderSource(source string, shaderType uint32, shader *uint32) bool {
	*shader = gl.CreateShader(shaderType)
	csource, free := gl.Strs(source)
	defer free()
	gl.ShaderSource(*shader, 1, csource, nil)
	gl.CompileShader(*shader)

	var compiled int32 = 0
	gl.GetShaderiv(*shader, gl.COMPILE_STATUS, &compiled)

	if compiled == 0 {
		var message [1024]uint8
		var messageSize int32 = 0
		gl.GetShaderInfoLog(*shader, 1024, &messageSize,
			(*uint8)(unsafe.Pointer(&message)))
		fmt.Fprintf(os.Stderr, "ERROR: could not compile %s\n", shaderTypeAsStr(shaderType))
		fmt.Fprintf(os.Stderr, "%.*s\n", messageSize, message)
		return false
	}

	return true
}

func compileShaderFile(filePath string, shaderType uint32, shader *uint32) bool {
	source := openFile(filePath)
	ok := compileShaderSource(source, shaderType, shader)
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: failed to compile `%s` shader file\n", filePath)
	}
	//free(source)
	return ok
}

func linkProgram(vertShader uint32, fragShader uint32, program *uint32) bool {
	*program = gl.CreateProgram()

	gl.AttachShader(*program, vertShader)
	gl.AttachShader(*program, fragShader)
	gl.LinkProgram(*program)

	var linked int32 = 0
	gl.GetProgramiv(*program, gl.LINK_STATUS, &linked)
	if linked == 0 {
		var message [1024]uint8
		var messageSize int32 = 0
		gl.GetProgramInfoLog(*program, 1024, &messageSize,
			(*uint8)(unsafe.Pointer(&message)))
		fmt.Fprintf(os.Stderr, "Program Linking: %.*s\n", messageSize, message)
		return false
	}

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)
	return true
}

type Uniform int

const (
	RESOLUTION_UNIFORM Uniform = iota
	TIME_UNIFORM
	MOUSE_UNIFORM
)

func (uniform Uniform) String() string {
	return [3]string{"resolution", "time", "mouse"}[uniform]
}

var programFailed, pause bool = false, false
var time, timeStep float64 = 0.0, 0.1
var mainProgram uint32 = 0
var mainUniforms [3]int32
var defaultWidth, defaultHeight int = 960, 540
var fragmentShaders []string
var currentShader int = 0

func loadShaderProgram(vertFilePath string, fragFilePath string, program *uint32) bool {
	var vert uint32 = 0
	if !compileShaderFile(vertFilePath, gl.VERTEX_SHADER, &vert) {
		return false
	}

	var frag uint32 = 0
	if !compileShaderFile(fragFilePath, gl.FRAGMENT_SHADER, &frag) {
		return false
	}

	if !linkProgram(vert, frag, program) {
		return false
	}
	return true
}

func updateShaderList() {
	fragmentShaders, _ = filepath.Glob(`./shaders/*.frag`)
	if fragmentShaders == nil {
		fmt.Fprintln(os.Stderr, "ERROR: no fragment shaders were found.")
		os.Exit(1)
	}
}

func reloadShaders() {
	gl.DeleteProgram(mainProgram)

	programFailed = true
	gl.ClearColor(1.0, 0.0, 0.0, 1.0)

	if !loadShaderProgram("./shaders/main.vert",
		fragmentShaders[currentShader], &mainProgram) {
		return
	}

	gl.UseProgram(mainProgram)

	for index := 0; index < len(mainUniforms); index++ {
		mainUniforms[index] = gl.GetUniformLocation(mainProgram, gl.Str(Uniform(index).String()+"\000"))
	}

	programFailed = false
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)

	fmt.Println("Successfully Reload the Shaders, current fragment shader:",
		fragmentShaders[currentShader])
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyF5:
			reloadShaders()
		case glfw.KeySpace:
			pause = !pause
		case glfw.KeyLeft:
			if pause {
				time -= timeStep
			}
		case glfw.KeyRight:
			if pause {
				time += timeStep
			}
		case glfw.KeyUp:
			updateShaderList()
			currentShader = (currentShader + 1) % len(fragmentShaders)
			reloadShaders()
		case glfw.KeyDown:
			updateShaderList()
			currentShader = (len(fragmentShaders) - 1 + currentShader) % len(fragmentShaders)
			reloadShaders()
		case glfw.KeyQ:
			os.Exit(0)
		}
	}
}

func windowSizeCallback(window *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not initialize GLFW, %s\n", err)
		os.Exit(1)
	}

	window, err := glfw.CreateWindow(defaultWidth, defaultHeight,
		"OpenGL", nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create GLFW window, %s\n",
			err)
		glfw.Terminate()
		os.Exit(1)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize GL, %s\n", err)
		os.Exit(1)
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	updateShaderList()
	reloadShaders()
	window.SetKeyCallback(keyCallback)
	window.SetFramebufferSizeCallback(windowSizeCallback)
	time = glfw.GetTime()
	var previousTime float64

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.PolygonMode(gl.FRONT_FACE, gl.FILL)

		if !programFailed {
			width, height := window.GetSize()
			gl.Uniform2f(mainUniforms[RESOLUTION_UNIFORM], float32(width),
				float32(height))
			gl.Uniform1f(mainUniforms[TIME_UNIFORM], float32(time))
			xpos, ypos := window.GetCursorPos()
			gl.Uniform2f(mainUniforms[MOUSE_UNIFORM], float32(xpos),
				float32(float64(height)-ypos))
			gl.DrawArraysInstanced(gl.TRIANGLE_STRIP, 0, 4, 1)
		}

		window.SwapBuffers()
		glfw.PollEvents()

		currentTime := glfw.GetTime()
		if !pause {
			time += currentTime - previousTime
		}
		previousTime = currentTime
	}
	os.Exit(0)
}
