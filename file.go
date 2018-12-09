package log

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
)

/**
 * DESCRIPTION:
 *
 * @author rd
 * @create 2018-12-08 17:11
 **/

func createDir(dir string) error {
    file, err := os.Open(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return os.MkdirAll(dir, os.ModePerm)
        }
        return err
    }
    var info os.FileInfo
    if info, err = file.Stat(); err != nil {
        return err
    }
    if info.IsDir() {
        return nil
    }
    return fmt.Errorf("the file:%s is not directory", dir)
}

func generateName(name, tag string, t time.Time) (string, string) {
    return fmt.Sprintf("%s_%s_%s.log", name, tag, t.Format("2006-01-02T15:04:05")), fmt.Sprintf("%s_%s.log", name, tag)
}

func createFile(name, dir, tag string, t time.Time) (*os.File, error) {
    err := createDir(dir)
    if err != nil {
        return nil, err
    }
    fileName, linkName := generateName(name, tag, t)
    fileName = filepath.Join(dir, fileName)
    file, err := os.Create(fileName)
    if err != nil {
        return nil, err
    }
    linkName = filepath.Join(dir, linkName)
    if err = os.Remove(linkName); err != nil {
        if os.IsNotExist(err) {
            goto label
        }
        file.Close()
        return nil, err
    }
label:
    if err = os.Link(fileName, linkName); err != nil {
        file.Close()
        return nil, err
    }
    return file, nil
}
