package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

func main() {
	// fmt.Printf("got args: %s", os.Args)
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s OBJ_NUM", os.Args[0])
		os.Exit(1)
	}

	objNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var outFile string
	if len(os.Args) == 3 {
		outFile = os.Args[2]
	}
	err = inspectPdfObject("input.pdf", objNum, outFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func inspectPdfObject(inputPath string, objNum int, outFile string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	if isEncrypted {
		// If encrypted, try decrypting with an empty one.
		// Can also specify a user/owner password here by modifying the line below.
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			fmt.Printf("Decryption error: %v\n", err)
			return err
		}
		if !auth {
			fmt.Println(" This file is encrypted with opening password. Modify the code to specify the password.")
			return nil
		}
	}

	obj, err := pdfReader.GetIndirectObjectByNumber(objNum)
	if err != nil {
		return err
	}

	fmt.Printf("Object %d: %s\n", objNum, obj.String())

	if stream, is := obj.(*core.PdfObjectStream); is {
		decoded, err := core.DecodeStream(stream)
		if err != nil {
			return err
		}
		if outFile != "" {
			f, err := os.Create(outFile)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.Write(decoded)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Decoded:\n%s", decoded)
		}
	} else if indObj, is := obj.(*core.PdfIndirectObject); is {
		fmt.Printf("%T\n", indObj.PdfObject)
		fmt.Printf("%s\n", indObj.PdfObject.String())
	}

	return nil
}
