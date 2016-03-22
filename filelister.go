package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "flag"
  "encoding/json"
  "time"
  "gopkg.in/yaml.v2"
)

// Necessary struct for JSON/YAML objects
type FileDir struct {
  ModifiedTime time.Time
  IsLink bool
  IsDir bool
  LinksTo string
  Size int64
  Name string
  Children []FileDir
}

// Prints out all files in plain text
func toText(path string, recursive bool, padding string) {

  // Loads files into memory
  files, error := ioutil.ReadDir(path)

  // Error checking
  if error != nil {
    fmt.Println("Error reading directory: " + path)
    return
  }

  // Prints out the name of each file
  for _, file := range files {

    // Checks if file is a symbolic link
    _, error := os.Readlink(path + file.Name())

    switch {

    // File is a directory to be printed recursively
    case file.IsDir() && recursive:

      // Prints the folder name indented with padding
      fmt.Println(padding + file.Name() + "/")

      // Recursively prints the folder's contents
      new_path := path + file.Name() + "/"
      toText(new_path, recursive, padding + "\t")

    // Nonrecursive folder
    case file.IsDir():
      fmt.Println(padding + file.Name() + "/")

    // Marks symbolic links with an asterisk
    case error == nil:
      fmt.Println(padding + file.Name() + "*")

    // Prints out filenames
    default:
      fmt.Println(padding + file.Name())
    }
  }
}

// Returns a FileDir(see above struct) array of all files in the path
func toFileDir(path string, recursive bool) []FileDir{
  files, err := ioutil.ReadDir(path)

  // Error checking
  if err != nil {
    fmt.Println("Error reading directory: " + path)
  }

  // The array (slice?) in which we will store the files
  obj := []FileDir{}

  // Prints out the name of each file
  for _, file := range files {

    // Checks if file is a symbolic link
    str, error := os.Readlink(path + file.Name())

    // Builds the struct
    fd := FileDir{
      ModifiedTime: file.ModTime(),
      IsLink: error == nil,
      IsDir:  file.Mode().IsDir(),
      LinksTo:  str,
      Size: file.Size(),
      Name: file.Name(),
      Children: []FileDir{},
    }

    // Populates children if recursive
    if recursive && fd.IsDir {
      new_path := path + file.Name() + "/"
      fd.Children = append(fd.Children, toFileDir(new_path, recursive)...)
    }

    // Appends the FileDir object to our array
    obj = append(obj, fd)
  }

  return obj
}

func main() {
  // Gathers data from CLI flags
  path := flag.String("path", "", "REQUIRED: The directory to print. If not sure, on a UNIX system try 'pwd' to get a valid path")
  output := flag.String("output", "text", "OPTIONAL: Output Type: 'text', 'json', or 'yaml'")
  recursive := flag.Bool("recursive", false, "OPTIONAL: When set, recursively displays a directory's contents")
  flag.Parse()

  // Checks path existence
  if *path == "" {
    fmt.Println("Flag -path is required.")
    return
  }

  // Checks that the path is valid
  _, er := ioutil.ReadDir(*path)
  if er != nil {
    fmt.Println("Invalid path/directory: " + *path)
    fmt.Println("If running on a unix-based system, command 'pwd' provides a valid path")
    return
  }

  // Adds a backslash to the end of the path if there isn't one
  if (*path)[len(*path)-1] != '/'{
    *path = *path + "/"
  }

  // Verifies output
  if (*output != "text") && (*output != "json") && (*output != "yaml") {
    fmt.Println("Invalid output type. Try 'text', 'json', or 'yaml'")
    return
  }

  // Prints out the path to be searched
  fmt.Println(*path)

  // Gets struct object
  fd := toFileDir(*path, *recursive)

  // Produces output according to the provided type
  switch {

  case *output == "json":

    // Converts data into JSON (pretty-printing)
    b, err := json.MarshalIndent(fd, "", "\t")

    // Error checking
    if err != nil {
      fmt.Println("Could not convert to JSON: ", err)
    }

    // Prints
    os.Stdout.Write(b)

  case *output == "yaml":

    // Converts data into YAML
    b, err := yaml.Marshal(fd)

    // Error checking
    if err != nil {
      fmt.Println("Could not convert to YAML: ", err)
    }

    // Prints
    os.Stdout.Write(b)

  // Text Output
  default:
    toText(*path, *recursive, "\t")
  }
}
