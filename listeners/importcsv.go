package listeners

// func importcsv(ip *insertParameters, db *sql.DB) {
// 	log.Println("Importing CSV file:", ip.csvfile)
// 	if f, err := os.Open(ip.csvfile); err == nil {
// 		scanner := bufio.NewScanner(f)
// 		scanner.Split(bufio.ScanLines)
// 		lineno := 0

// 		b := initBatch()
// 		for scanner.Scan() {
// 			lineno = lineno + 1
// 			if lineno == 1 {
// 				continue
// 			}

// 			values := strings.Split(scanner.Text(), ",")
// 			if len(values) == 4 {
// 				parts := strings.Split(values[0], "\\")
// 				name := parts[len(parts)-1]
// 				time, err := time.Parse("2006-01-02 15:04:05", values[1])
// 				if err != nil {
// 					log.Println("Failed to parse time:", values[1], err)
// 					continue
// 				}

// 				value, err := strconv.ParseFloat(values[2], 64)
// 				if err != nil {
// 					log.Println("Failed to convert value to float64:", values[2], err)
// 					continue
// 				}

// 				quality, err := strconv.Atoi(values[3])
// 				if err != nil {
// 					log.Println("Failed to convert quality to int:", values[3], err)
// 					continue
// 				}

// 				dp := &DataPoint{Time: time, Name: name, Value: value, Quality: quality}
// 				if full := b.appendPoint(dp); full {
// 					b.insertBatch(db)
// 					b = initBatch()
// 				}
// 			}
// 		}

// 		b.insertBatch(db)

// 	} else {
// 		log.Println("Failed to open file:", ip.csvfile)
// 	}
// }
