package fsutil

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCheckIfCompressed(t *testing.T) {
	var compressedZipFile = "./TestFiles/ZipFiles/test.zip"
	var uncompressedZipFile = "./TestFiles/OtherFiles/othertestfile.txt"

	var tests = []struct {
		source       string
		shouldAssert bool
	}{
		{compressedZipFile, true},
		{uncompressedZipFile, false},
	}

	for _, test := range tests {
		f, err := os.Open(test.source)
		if err != nil {
			t.Error("Error opening file", err)
		}
		r := CheckIfCompressed(f)
		if r != test.shouldAssert {
			t.Errorf("Test should of returned %v but returned %v instead", test.shouldAssert, r)
		}
	}

}

func TestCopyFile(t *testing.T) {
	var badSourceFile = "./TestFiles/FGFailedSourceFileLocation.fg"
	var badDestinationFile = "./TestFiles/NonexistantDirectory/FGFailedDestFileLocation.fg"
	var goodSourceFile = "./TestFiles/OtherFiles/othertestfile.txt"
	var goodDestinationFile = "./TestFiles/OtherFiles/othertestfile_copy.txt"
	var symlinkSourceFile = "./TestFiles/OtherFiles/TestSubdirectory/TextFiles"
	var symlinkDestinationFile = "./TestFiles/OtherFiles/TestSubdirectory/TextFilesCopy"

	var tests = []struct {
		source       string
		destination  string
		shouldAssert bool
	}{
		{"", goodDestinationFile, false},
		{goodSourceFile, "", false},
		{badSourceFile, goodDestinationFile, false},
		{goodSourceFile, badDestinationFile, false},
		{goodSourceFile, goodDestinationFile, true},
		{symlinkSourceFile, symlinkDestinationFile, true},
	}

	for index, test := range tests {
		err := CopyFile(test.source, test.destination)
		if err != nil && test.shouldAssert == true {
			t.Errorf("Test %d failed but should have passed. "+err.Error(), index)
		} else if err == nil && test.shouldAssert == false {
			t.Errorf("Test %d passed but should have failed.", index)
		} else if err == nil && test.shouldAssert == true {
			if _, err := os.Stat(test.source); err != nil {
				t.Error("CopyFile did not throw an error, but the test file was not created.")
			}

			if err := os.Remove(test.destination); err != nil {
				t.Error("Unable to clean up after running CopyFile tests. " + err.Error())
			}
		}
	}
}

func TestCopyDirectory(t *testing.T) {
	var badSourceDirectory = "./TestFiles/BadSourceFolder"
	var goodSourceDirectory = "./TestFiles/OtherFiles"
	var goodDestinationDirectory = "./TestFiles/TextFilesCopy"

	var tests = []struct {
		source       string
		destination  string
		shouldAssert bool
	}{
		{"", goodDestinationDirectory, false},
		{goodSourceDirectory, "", false},
		{badSourceDirectory, goodDestinationDirectory, false},
		{goodSourceDirectory, goodDestinationDirectory, true},
	}

	var testsPassed = true
	for index, test := range tests {
		err := CopyDirectory(test.source, test.destination, true)
		if err != nil && test.shouldAssert == true {
			t.Errorf("Test %d failed but should have passed.", index)
			testsPassed = false
		} else if err == nil && test.shouldAssert == false {
			t.Errorf("Test %d passed but should have failed.", index)
			testsPassed = false
		}
	}

	if testsPassed {
		if _, err := os.Stat(goodDestinationDirectory); err != nil {
			t.Error("CopyDirectory test directory was not created.")
		}
	}

	if err := os.RemoveAll(goodDestinationDirectory); err != nil {
		t.Error("Unable to clean up after running CopyDirectory tests. " + err.Error())
	}
}

func TestIsEmpty(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error("Error creating temp file", err)
	}

	tests := []struct {
		dir      string
		expected bool
		pass     bool
	}{
		{"", false, false},
		{dir, true, true},
		{"TestFiles", false, true},
	}

	for _, test := range tests {
		result, err := IsEmpty(test.dir)
		if err != nil && test.pass {
			t.Error("IsEmpty errored", err)
		} else if err == nil && !test.pass {
			t.Error("IsEmpty should have erroed")
		}

		if result != test.expected {
			t.Errorf("Expected %t, Result was: %t", test.expected, result)
		}
	}

	// Remove temp dir
	os.RemoveAll(dir)
}

func TestRemoveDirContents(t *testing.T) {
	tests := []struct {
		testDir string
		tempDir string
		pass    bool
	}{
		{"./TestFiles/TextFiles", "./TestFiles/TextFilesCopy", true},
		{"", "", false},
	}

	for _, test := range tests {
		// Make a copy of the directory for testing
		if test.pass {
			err := CopyDirectory(test.testDir, test.tempDir, true)
			if err != nil {
				t.Error("Unabled to copy directory", err)
			}
		}

		// Remove all contents in TextFilesCopy
		err := RemoveDirContent(test.tempDir)
		if err != nil && test.pass {
			t.Error("Error Removing Directory Contents", err)
		} else if err == nil && !test.pass {
			t.Error("RemoveDirContents should have errored")
		}

		// Test if directory is empty
		empty, err := IsEmpty(test.tempDir)
		if err != nil && test.pass {
			t.Error("IsEmpty errored")
		} else if err == nil && !test.pass {
			t.Error("IsEmpty should have errored")
		}

		if !empty && test.pass {
			t.Error("RemoveDirContents Failed. Directory should be empty")
		}

		// Delete TextFilesCopy
		os.RemoveAll(test.tempDir)
	}

}
