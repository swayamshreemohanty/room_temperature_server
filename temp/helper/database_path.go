package helper


type Path string

var DatabasePath = struct {
    DATABASE         Path
    STATUS		 Path
}{
    DATABASE: "HomeAuto",
    STATUS: "status",

}