package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
)

type Params struct {
	ImagePath string
	Port      int
}

func main() {
	params := ParseParameters()
	imageConfig := LoadImage(params.ImagePath)
	imageMask := CreateImageMask(*imageConfig)
	s := Server{
		ImageConfig: imageConfig,
		ImageMask:   imageMask,
		Port:        params.Port,
	}
	InitServer(s)
}

// Pass CLI args
func ParseParameters() *Params {
	image := flag.String("imagePath", "", "The path to the image to display")
	port := flag.Int("port", 8080, "The port to run the app on")
	flag.Parse()

	if *image == "" {
		log.Fatal("the image path must be set")
	}

	params := Params{
		ImagePath: *image,
		Port:      *port,
	}

	return &params
}

type ImageConfig struct {
	Width     int
	Height    int
	ImageData []byte
	Path      string
	Format    string
}

// Attempt to load image file and return raw data
func LoadImage(path string) *ImageConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("unable to load image file: ", err)
	}

	f, _ := os.Open(path)
	config, format, err := image.DecodeConfig(f)
	if err != nil {
		log.Fatal("could not decode config, ", err)
	}
	ic := ImageConfig{
		Width:     config.Width,
		Height:    config.Height,
		Format:    format,
		ImageData: data,
		Path:      path,
	}
	log.Printf("Image loaded with dimensions %dw x %dh, format: %s", ic.Width, ic.Height, ic.Format)
	return &ic
}

// Create an alpha image the same size as the source image, that is filled with opacity
func CreateImageMask(ic ImageConfig) *image.Alpha {
	bounds := image.Rectangle{
		Min: image.Pt(0, 0),
		Max: image.Pt(ic.Width, ic.Height),
	}
	mask := image.NewAlpha(bounds)
	for x := 0; x < ic.Width; x++ {
		for y := 0; y < ic.Height; y++ {
			mask.SetAlpha(x, y, color.Alpha{A: 255})
		}
	}
	return mask
}

type Server struct {
	ImageConfig *ImageConfig
	ImageMask   *image.Alpha
	Port        int
}

// Initialise web server and routes
func InitServer(s Server) {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./client"))
	mux.Handle("/", fs)
	mux.HandleFunc("/join", s.JoinHandler)
	mux.HandleFunc("/image", s.ImageHandler)
	mux.HandleFunc("/mask", s.MaskHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux))
}

func (s *Server) JoinHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Join"))
}

func (s *Server) ImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", s.ImageConfig.Format))
	w.Write(s.ImageConfig.ImageData)
}

func (s *Server) MaskHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create(fmt.Sprintf("mask.%s.png", s.ImageConfig.Path))
	if err != nil {
		log.Println("can't create mask file, ", err)
	}
	defer f.Close()
	png.Encode(f, s.ImageMask)
	var image []byte
	_, err = f.Read(image)
	if err != nil {
		log.Println("can't read created mask, ", err)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(image)
}
