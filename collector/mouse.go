package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Mouse struct {
	Name     string
	BtnCount int
	VWheel   bool
	HWheel   bool
}

// MouseDevice Структура для хранения информации о мыши
type MouseDevice struct {
	Name string
	KEY  string
	REL  string
}

//const INPUT_DEVICES = "/Users/kotlinist/workspace/projects/golang/collector/fake-devices" // /proc/bus/input/devices

// Функция для извлечения всех устройств мыши из файла /proc/bus/input/devices
func getMouseDevices(path string) ([]MouseDevice, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer file.Close()

	// Регулярные выражения для поиска данных
	nameRegex := regexp.MustCompile(`^N: Name="(.+)"`)            // Имя устройства
	handlersRegex := regexp.MustCompile(`^H: Handlers=.*mouse\d`) // Устройство мыши
	keyRegex := regexp.MustCompile(`^B: KEY=(.+)`)                // Строка с KEY=
	relRegex := regexp.MustCompile(`^B: REL=(.+)`)                // Строка с REL=

	var mice []MouseDevice
	var currentMouse *MouseDevice

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Если строка содержит имя устройства
		if matches := nameRegex.FindStringSubmatch(line); matches != nil {
			currentMouse = &MouseDevice{
				Name: matches[1],
			}
		}

		// Если строка указывает, что это мышь
		if matches := handlersRegex.FindStringSubmatch(line); matches != nil && currentMouse != nil {
			// Добавляем устройство мыши в список
			mice = append(mice, *currentMouse)
			currentMouse = &mice[len(mice)-1]
		}

		// Если строка содержит KEY= и мы в разделе мыши
		if matches := keyRegex.FindStringSubmatch(line); matches != nil && currentMouse != nil {
			currentMouse.KEY = matches[1]
		}

		// Если строка содержит REL= и мы в разделе мыши
		if matches := relRegex.FindStringSubmatch(line); matches != nil && currentMouse != nil {
			currentMouse.REL = matches[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла: %v", err)
	}

	return mice, nil
}

// HasScrollWheels принимает битовую маску REL и возвращает логические значения
// наличия вертикального и горизонтального колес прокрутки.
func hasScrollWheels(mask string) (hasHorizontal, hasVertical bool) {
	// Предположим, что:
	// Бит 0 (1 << 0) отвечает за горизонтальное колесо прокрутки.
	// Бит 1 (1 << 1) отвечает за вертикальное колесо прокрутки.
	REL, err := strconv.Atoi(mask)
	if err != nil {
		//fmt.Println(err)
		return false, false
	}

	hasVertical = REL&(1<<3) != 0   // Проверка наличия вертикального колеса
	hasHorizontal = REL&(1<<8) != 0 // Проверка наличия горизонтального колеса

	return
}

func getMouseInfo() []Mouse {
	devices, err := getMouseDevices(inputDevicesFile)
	if err != nil {
		return nil
	}
	var mice []Mouse
	for i := range devices {
		//fmt.Println(devices[i])
		mouse := Mouse{}
		hasVerticalWheel, hasHorizontalWheel := hasScrollWheels(devices[i].REL)
		mouse.Name = devices[i].Name
		mouse.VWheel = hasVerticalWheel
		mouse.HWheel = hasHorizontalWheel
		mouse.BtnCount = countSetBits(devices[i].KEY)
		mice = append(mice, mouse)
	}
	return mice
}

// Функция подсчёта установленных битов в битовой маске
func countSetBits(keyMask string) int {
	// Разбиваем битовую маску на части (по пробелам)
	parts := strings.Fields(keyMask)
	count := 0

	for _, part := range parts {
		// Преобразуем шестнадцатеричную строку в число
		num, err := strconv.ParseUint(part, 16, 64)
		if err != nil {
			continue
		}
		// Подсчитываем установленные биты
		count += popCount(num)
		//fmt.Printf("count=%d\n, num=%d", count, num)
	}

	return count
}

// Функция подсчёта установленных битов в числе
func popCount(x uint64) int {
	count := 0
	for x > 0 {
		count += int(x & 1)
		x >>= 1
	}
	return count
}
