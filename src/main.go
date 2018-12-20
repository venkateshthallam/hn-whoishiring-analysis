package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/caser/gophernews"
	"github.com/gocolly/colly"
)

type comment struct {
	Author  string `selector:"a.hnuser"`
	URL     string `selector:".age a[href]" attr:"href"`
	Comment string `selector:".comment"`
	Replies []*comment
	depth   int
}

// Locations ..
type Locations struct {
	City  string  `json:"city"`
	Lat   float64 `json:"lat"`
	Long  float64 `json:"long"`
	Count int     `json:"count"`
}

// JobPosts ..
type JobPosts struct {
	CompanyName string   `json:"company_name"`
	Location    string   `json:"location"`
	JobType     string   `json:"job_type"`
	Salary      string   `json:"salary"`
	Skills      []string `json:"skills"`
	Visa        string   `json:"visa"`
	RemoteWork  string   `json:"remote_work"`
}

var titles = map[string]int{
	"Software Engineer":           0,
	"Senior Software Engineer":    0,
	"Product Manager":             0,
	"Program Manager":             0,
	"Engineering Manager":         0,
	"Staff Software Engineer":     0,
	"Principal Software Engineer": 0,
	"Product Designer":            0,
	"QA":                          0,
}

var locations = map[string]int{
	"New York":       0,
	"San Francisco":  0,
	"Los Angeles":    0,
	"Boston":         0,
	"Austin":         0,
	"Dallas":         0,
	"Denver":         0,
	"Seattle":        0,
	"NYC":            0,
	"San Jose":       0,
	"San Diego":      0,
	"Salt Lake City": 0,
	"Portland":       0,
	"Kansas City":    0,
}

// skills from stack over flow tags
// var skills = map[string]int{
// 	"scala":       0,
// 	"awk":         0,
// 	"julia":       0,
// 	"rust":        0,
// 	"haskell":     0,
// 	"python":      0,
// 	"java":        0,
// 	"javascript ": 0,
// 	"typescript":  0,
// 	"golang":      0,
// 	"ruby":        0,
// 	"perl":        0,
// 	"shell":       0,
// 	"kubernetes":  0,
// 	"rails":       0,
// 	"django":      0,
// 	"spring boot": 0,
// 	"graphql":     0,
// 	"lua":         0,
// 	"elixir":      0,
// 	"erlang":      0,
// 	"kotlin":      0,
// 	"d":           0,
// 	"docker":      0,
// 	"spring":      0,
// 	"hibernate":   0,
// 	"android":     0,
// 	"ios":         0,
// 	"swift":       0,
// 	"php":         0,
// 	"css":         0,
// 	"c#":          0,
// 	".net":        0,
// 	"html":        0,
// 	"c":           0,
// 	"c\\+\\+":     0,
// 	"mysql":       0,
// 	"postgres":    0,
// 	"sql":         0,
// 	"objective-c": 0,
// 	"asp.net":     0,
// 	"angular":     0,
// 	"angularjs":   0,
// 	"react":       0,
// 	"reactjs":     0,
// 	"react.js":    0,
// 	"vuejs":       0,
// 	"vue.js":      0,
// 	"sql-server":  0,
// 	"ajax":        0,
// 	"excel":       0,
// 	"linux":       0,
// 	"html5":       0,
// 	"git":         0,
// 	"apache":      0,
// 	"matlab":      0,
// 	"f#":          0,
// }

var skills = map[string]int{
	"\\bscala\\b":       0,
	"\\bawk\\b":         0,
	"\\bjulia\\b":       0,
	"\\brust\\b":        0,
	"\\bhaskell\\b":     0,
	"\\bpython\\b":      0,
	"\\bjava\\b":        0,
	"\\bjavascript \\b": 0,
	"\\btypescript\\b":  0,
	"\\bgolang\\b":      0,
	"\\bruby\\b":        0,
	"\\bperl\\b":        0,
	"\\bshell\\b":       0,
	"\\bkubernetes\\b":  0,
	"\\brails\\b":       0,
	"\\bdjango\\b":      0,
	"\\bspring boot\\b": 0,
	"\\bgraphql\\b":     0,
	"\\blua\\b":         0,
	"\\belixir\\b":      0,
	"\\berlang\\b":      0,
	"\\bkotlin\\b":      0,
	"\\bd\\b":           0,
	"\\bdocker\\b":      0,
	"\\bspring\\b":      0,
	"\\bhibernate\\b":   0,
	"\\bandroid\\b":     0,
	"\\bios\\b":         0,
	"\\bswift\\b":       0,
	"\\bphp\\b":         0,
	"\\bcss\\b":         0,
	"\\bc#\\b":          0,
	"\\b.net\\b":        0,
	"\\bhtml\\b":        0,
	"\\bc\\b":           0,
	"\\bc\\+\\+\\b":     0,
	"\\bmysql\\b":       0,
	"\\bpostgres\\b":    0,
	"\\bsql\\b":         0,
	"\\bobjective-c\\b": 0,
	"\\basp.net\\b":     0,
	"\\bangular\\b":     0,
	"\\bangularjs\\b":   0,
	"\\breact\\b":       0,
	"\\breactjs\\b":     0,
	"\\breact.js\\b":    0,
	"\\bvuejs\\b":       0,
	"\\bvue.js\\b":      0,
	"\\bsql-server\\b":  0,
	"\\bajax\\b":        0,
	"\\bexcel\\b":       0,
	"\\blinux\\b":       0,
	"\\bhtml5\\b":       0,
	"\\bgit\\b":         0,
	"\\bapache\\b":      0,
	"\\bmatlab\\b":      0,
	"\\bf#\\b":          0,
}

// itemMonthMap maps moth to respective HN thread id
var itemMonthMap = map[int]int{
	1:  16052538,
	2:  16282819,
	3:  16492994,
	4:  16735011,
	5:  16967543,
	6:  17205865,
	7:  17442187,
	8:  17663077,
	9:  17902901,
	10: 18113144,
	11: 18354503,
	12: 18589702,
}

// Result ..
type Result struct {
	Parameter     string `json:"parameter"`
	Count         int    `json:"count"`
	Month         int    `json:"month"`
	CommentsCount int    `json:"comments_count"`
}

// saveResultsToFile ..
func saveJSStuffToFile(latLongMap map[string]*LatLong, mapData []*MapData) {
	resultsBlob, _ := json.Marshal(latLongMap)
	err := ioutil.WriteFile("lat_long_map.json", resultsBlob, 0644)
	checkErr(err)

	titleBlob, _ := json.Marshal(mapData)
	err = ioutil.WriteFile("map_data.json", titleBlob, 0644)
	checkErr(err)
}

func saveSkillsToFile(skillsMap map[string]int) {
	skillsJSON, _ := json.Marshal(skillsMap)
	err := ioutil.WriteFile("skills.json", skillsJSON, 0644)
	checkErr(err)
}

// saveResultsToFile ..
func saveResultsToFile(results []*Result, titleMap map[string]int) {
	resultsBlob, _ := json.Marshal(results)
	err := ioutil.WriteFile("results.json", resultsBlob, 0644)
	checkErr(err)

	titleBlob, _ := json.Marshal(titleMap)
	err = ioutil.WriteFile("title_results.json", titleBlob, 0644)
	checkErr(err)
}

// saveResultsToFile ..
func saveLocationsToFile(locations []*Locations) {
	resultsBlob, _ := json.Marshal(locations)
	err := ioutil.WriteFile("locations.json", resultsBlob, 0644)
	checkErr(err)
}

func getResultsAndPrint() {
	results := make([]*Result, 0)
	jsonFile, err := os.Open("/Users/vthallam/Documents/code/scraping/HN Comments/results.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &results)
	checkErr(err)

	visaMap := make(map[int]int, 0)
	remoteMap := make(map[int]int, 0)
	lunchMap := make(map[int]int, 0)
	for _, result := range results {
		if result.Parameter == "visa" {
			visaMap[result.Month] = result.Count
		}
		if result.Parameter == "remote" {
			remoteMap[result.Month] = result.Count
		}
		if result.Parameter == "lunch" {
			lunchMap[result.Month] = result.Count
		}
	}

	printMap(visaMap)
	fmt.Println()
	printMap(remoteMap)
	fmt.Println()
	printMap(lunchMap)

}

func printMap(pmap map[int]int) {
	for _, v := range pmap {
		fmt.Print(v, ", ")
	}
}

func getTextFromFiles() {
	file, err := os.Open("")
	checkErr(err)

	defer file.Close()

	lines := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
}

func main() {

	//gets all comments from who is hiring threads
	scrapeAllMonthsComments()

	// analyzes remote, visa and perks from comment text and saves as a result object
	totalResults := make([]*Result, 0)
	skillsCount := make(map[string]int, 0)
	for i, item := range itemMonthMap {
		itemStr := strconv.Itoa(item)
		results, _ := analyzeComments(itemStr+".json", i)
		analyzeSkills(itemStr+".json", i, skillsCount)

		aggregateTitleCount(itemStr+".json", i)
		totalResults = append(totalResults, results...)
	}
	saveSkillsToFile(skillsCount)
	saveResultsToFile(totalResults, titles)

	//helper methods to print the output in a way charts can understand
	getTitles()
	getResultsAndPrint()
	getLocations()
	gatherLatLong()
	getResultsAndPrint()

	// sorts skills by value and prints them which are used to for highcharts
	sortSkills()
}

func sortSkills() {
	type kv struct {
		Key   string
		Value int
	}

	skillMap := getSkills()
	var ss []kv
	for k, v := range skillMap {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	keyList := make([]string, 0)
	valueList := make([]int, 0)
	for _, kv := range ss {
		keyList = append(keyList, kv.Key)
		valueList = append(valueList, kv.Value)
		fmt.Printf("%s, %d\n", kv.Key, kv.Value)
	}

	fmt.Printf("key list ======== %+v \n", keyList)
	fmt.Printf(" valueList ======== %+v \n", valueList)

}

// getAllCommentsForItemID ..
func getAllCommentsForItemID(itemID int) []*comment {
	fmt.Println("getting comments for itemID ", itemID)
	client := gophernews.NewClient()
	story, err := client.GetStory(itemID)
	comments := make([]*comment, 0)
	if err != nil {
		panic(err)
	}

	for _, commentID := range story.Kids {
		gComment, err := client.GetComment(commentID)
		logErr(err)
		com := &comment{}
		com.Comment = gComment.Text
		comments = append(comments, com)
	}

	return comments
}

// LatLong ..
type LatLong struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// MapData..
type MapData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

func getLocations() {
	jsonFile, err := os.Open("locations.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	locations := make([]*Locations, 0)
	json.Unmarshal([]byte(byteValue), &locations)

	locLatLongMap := make(map[string]*LatLong, 0)
	mapdata := make([]*MapData, 0)
	for _, location := range locations {
		latlong := &LatLong{}
		md := &MapData{}
		latlong.Latitude = location.Lat
		latlong.Longitude = location.Long
		locLatLongMap[location.City] = latlong

		md.ID = location.City
		md.Name = location.City
		md.Value = location.Count
		md.Color = `chart.colors.getIndex(0)`
		mapdata = append(mapdata, md)
	}

	latlongJSON, _ := json.Marshal(locLatLongMap)
	mapDataJSON, _ := json.Marshal(mapdata)

	fmt.Println(latlongJSON)
	fmt.Println(mapDataJSON)

	saveJSStuffToFile(locLatLongMap, mapdata)
}

func getSkills() map[string]int {
	skillMap := make(map[string]int, 0)
	// Open our jsonFile
	jsonFile, err := os.Open("skills.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &skillMap)

	return skillMap
}

func getTitles() {
	locMap := make(map[string]int, 0)
	// Open our jsonFile
	jsonFile, err := os.Open("title_results.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &locMap)

	for title := range titles {
		fmt.Print(title, ", ")
	}
	fmt.Println()

	for _, v := range locMap {
		fmt.Print(v, ", ")
	}

}

func gatherLatLong() {
	locMap := make(map[string]int, 0)
	// Open our jsonFile
	jsonFile, err := os.Open("cities.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &locMap)

	locations := make([]*Locations, 0)
	for city, count := range locMap {
		location := &Locations{}
		location.City = city
		location.Count = count
		locations = append(locations, location)
	}

	for _, location := range locations {
		time.Sleep(1 * time.Second)
		postLocationData(location)
	}

	saveLocationsToFile(locations)
}

func postLocationData(location *Locations) {
	req, err := http.NewRequest("GET", "https://maps.googleapis.com/maps/api/place/findplacefromtext/json", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("key", "") // you need a API key from google cloud for this to work.
	q.Add("input", location.City)
	q.Add("inputtype", "textquery")
	q.Add("fields", "photos,formatted_address,name,rating,opening_hours,geometry")
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	checkErr(err)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	parseJSONAndAssignLatLong(body, location)
}

func parseJSONAndAssignLatLong(body []byte, location *Locations) {
	arr, _, _, _ := jsonparser.Get(body, "candidates")
	jsonparser.ArrayEach(arr, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		lat, _ := jsonparser.GetFloat(value, "geometry", "location", "lat")
		long, _ := jsonparser.GetFloat(value, "geometry", "location", "lng")
		location.Lat = lat
		location.Long = long
	})

	return
}

func scrapeAllMonthsComments() {
	//	wg := &sync.WaitGroup{}
	//wg.Add(12)
	for _, itemID := range itemMonthMap {
		comments := getAllCommentsForItemID(itemID)
		fileName := strconv.Itoa(itemID)
		writeCommentsToFile(fileName+".json", comments)
		time.Sleep(2 * time.Second)
		//scrapeCommentsForItemID(itemID, wg)
	}
	//wg.Wait()
}

func writeCommentsToFile(fileName string, comments []*comment) {
	commentsJSON, _ := json.Marshal(comments)
	err := ioutil.WriteFile(fileName, commentsJSON, 0644)
	if err != nil {
		panic(err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func logErr(err error) {
	if err != nil {
		fmt.Println(" err occured ", err)
	}
}

func aggregateTitleCount(fileName string, month int) {
	comments := getComments(fileName)
	for _, comment := range comments {
		commentText := comment.Comment
		for title := range titles {
			titles[title] += returnOneIfFound(title, commentText)
		}
	}
}

func analyzeComments(fileName string, month int) ([]*Result, map[string]int) {
	comments := getComments(fileName)
	remote := 0
	visa := 0
	lunch := 0
	fullTime := 0
	contractor := 0

	for _, comment := range comments {
		commentText := comment.Comment

		remote += returnOneIfFound("remote", commentText)
		visa += returnOneIfFound("visa", commentText)
		lunch += returnOneIfFound("lunch", commentText)
		fullTime += returnOneIfFound("Full Time", commentText)
		contractor += returnOneIfFound("Contractor", commentText)

	}
	visaResult := &Result{
		Parameter:     "visa",
		Count:         visa,
		CommentsCount: len(comments),
		Month:         month,
	}

	remoteResult := &Result{
		Parameter:     "remote",
		Count:         remote,
		CommentsCount: len(comments),
		Month:         month,
	}

	lunchResult := &Result{
		Parameter:     "lunch",
		Count:         lunch,
		CommentsCount: len(comments),
		Month:         month,
	}

	fmt.Println("most popular titles are --- ", titles)
	fmt.Println("most popular locations are --- ", locations)
	fmt.Println("no of companies which offer remote work -> ", remote)
	fmt.Println("no of companies which offer lunch at work -> ", lunch)
	fmt.Println("no of companies which sponsor visa -> ", visa)
	fmt.Println("no of companies which with full time jobs -> ", fullTime)
	fmt.Println("no of companies which hires contractors -> ", contractor)

	return []*Result{visaResult, remoteResult, lunchResult}, titles
}

func analyzeSkills(fileName string, month int, skillsCount map[string]int) {
	comments := getComments(fileName)

	//regex := "\\b%s\\b"
	//fmt.Println(regexp.MatchString("\\bews\\b", word))
	newmap := make(map[string]int, 0)
	count := 0
	for _, comment := range comments {
		count++
		commentText := comment.Comment

		for skill := range skills {
			if _, ok := newmap[skill]; !ok {
				newmap[skill] = 0
			}
			//pattern := fmt.Sprintf(regex, skill)
			//fmt.Println("pattern is ", pattern)
			newmap[skill] += returnOneIfFound(skill, commentText)
		}
	}
	fmt.Println("newmap =====> ", newmap)
	fmt.Println("number of comments processed ", count)
	//fmt.Println("most popular titles are --- ", skills)
}

func returnOneIfFound(pattern string, text string) int {
	found, err := regexp.MatchString(pattern, text)
	checkErr(err)
	if found {
		return 1
	}
	return 0
}

func getComments(fileName string) []*comment {
	// Open our jsonFile
	jsonFile, err := os.Open("/Users/vthallam/Documents/code/scraping/HN Comments/" + fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened ", fileName)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	comments := make([]*comment, 0)
	err = json.Unmarshal([]byte(byteValue), &comments)
	if err != nil {
		panic(err)
	}

	return comments
}

// scrapeAndStoreHiringComments method uses Colly, a golang scraper to scrape comments.
func scrapeAndStoreHiringComments(itemID string) []*comment {

	comments := make([]*comment, 0)

	// Instantiate default collector
	c := colly.NewCollector()

	// Extract comment
	c.OnHTML(".comment-tree tr.athing", func(e *colly.HTMLElement) {
		width, err := strconv.Atoi(e.ChildAttr("td.ind img", "width"))
		if err != nil {
			return
		}
		// hackernews uses 40px spacers to indent comment replies,
		// so we have to divide the width with it to get the depth
		// of the comment
		depth := width / 40
		c := &comment{
			Replies: make([]*comment, 0),
			depth:   depth,
		}
		e.Unmarshal(c)
		//fmt.Println(c.Comment, "\n ============= ", len(c.Comment))
		if len(c.Comment) > 5 {
			c.Comment = strings.TrimSpace(c.Comment[:len(c.Comment)-5])
		}
		if depth == 0 {
			comments = append(comments, c)
			return
		}
		parent := comments[len(comments)-1]
		// append comment to its parent
		for i := 0; i < depth-1; i++ {
			parent = parent.Replies[len(parent.Replies)-1]
		}
		parent.Replies = append(parent.Replies, c)
	})

	c.Visit("https://news.ycombinator.com/item?id=" + itemID)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	return comments
}
