# fsutil
File system utility functions

## Install
`go get -u github.com/forestgiant/fsutil`

## Usage
* `CheckIfCompressed`
  * file must be io.ReadSeeker: Reads the first 512 bytes of a file into a buffer to check http.DetectContentType. After the read it seeks file back to start.
* `Unzip`
  * Takes a src compressed file and unzips the contents to the destination file.
* `FileExists`
  * Test to see if a file exists and return true or false
* `CopyFile`
  * Copies the file at the source path to the provided destination.
* `CopyDirectory`
  * Copies the directory at the source path to the provided destination, with the option of recursively copying subdirectories.
  
