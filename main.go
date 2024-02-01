package main

import (
	"bytes"
	"fmt"
	"github.com/bitfield/script"
	"os"
	"strconv"
	"strings"
	"time"
)

// read drawing
// nothing on ., commit on x
// make drawing matrix
// https://stackoverflow.com/questions/25563455/how-do-i-get-last-commit-date-from-git-repository
// get last commit date
// make the necessary commits until today
// save x,y on file
// wrap around on end
// top == sunday
func main() {
	dataRaw, err := os.ReadFile("./drawing.txt")
	if err != nil {
		panic(err)
	}
	lines := bytes.Split(dataRaw, []byte("\n"))
	mat := make([][]byte, 0, len(lines))
	for x, line := range lines {
		mat = append(mat, make([]byte, 0, len(line)))
		for _, ch := range line {
			mat[x] = append(mat[x], ch)
		}
	}
	for _, line := range mat {
		for _, ch := range line {
			fmt.Print(string(ch))
		}
		fmt.Println()
	}
	for y := 0; y < len(mat[0]); y++ {
		for x := 0; x < len(mat); x++ {
			fmt.Print(string(mat[x][y]))
		}
		fmt.Println()
	}
	currX, currY, err := getState()
	if err != nil {
		panic(err)
	}
	fmt.Println("x", currX, "y", currY)
	lastDate, err := getLastCommitDate()
	if err != nil {
		panic(err)
	}
	pointsToCommit, err := daysAfterToCommit(mat, currX, currY, lastDate)
	fmt.Printf("%#v, %#v\n", pointsToCommit, err)
	err = processPoints(pointsToCommit, lastDate)
}

func getState() (x, y int, err error) {
	data, err := os.ReadFile("./current.txt")
	if err != nil {
		return -1, -1, err
	}
	splt := bytes.Split(data, []byte(","))
	x, err = strconv.Atoi(string(splt[0]))
	if err != nil {
		return -1, -1, err
	}
	y, err = strconv.Atoi(strings.Trim(string(splt[1]), "\n\t "))
	if err != nil {
		return -1, -1, err
	}
	return x, y, nil
}

func getLastCommitDate() (time.Time, error) {
	dt, err := script.Exec("git log -1 --date=format:\"%Y-%m-%d\" --format=\"%ad\"").String()
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.DateOnly, dt[:10])
}

func daysAfterToCommit(mat [][]byte, currX, currY int, lastCommitDate time.Time) ([][]int, error) { //days to commit, and last x and y
	outToCommit := [][]int{}
	today, err := time.Parse(time.DateOnly, time.Now().Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	dayPtr := lastCommitDate
	count := 0
	x, y := currY, currX
	fmt.Println(len(mat), len(mat[0]))
	for {
		count++
		dayPtr = dayPtr.Add(time.Hour * 24)
		if dayPtr.After(today) {
			break
		}
		if mat[x][y] == 'x' {
			outToCommit = append(outToCommit, []int{count, x, y})
		}
		fmt.Print(string(mat[x][y]))
		x++
		if x == 7 {
			x = 0
			y++
			if y == len(mat[0]) {
				y = 0
			}
			fmt.Println()
		}
	}
	return outToCommit, nil
}

func processPoints(points [][]int, lastCommit time.Time) error {
	for _, p := range points {
		err := os.WriteFile("./current.txt", []byte(fmt.Sprintf("%d,%d", p[1], p[2])), 777)
		if err != nil {
			return err
		}
		_, err = script.Exec("git add current.txt").Stdout()
		if err != nil {
			return err
		}
		_, err = script.Exec(fmt.Sprintf("git commit --date='%s' -m '%v'", lastCommit.Add(time.Hour*24*time.Duration(p[0])).Format(time.DateOnly), lastCommit.Add(time.Hour*24*time.Duration(p[0])).Format(time.DateOnly))).Stdout()
		if err != nil {
			return err
		}
	}
	return nil
}
