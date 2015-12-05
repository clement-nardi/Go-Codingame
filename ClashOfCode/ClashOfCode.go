//Pyramid
package main
import "fmt"
func main() {
	var n int
	fmt.Scan(&n)
	for i := 0; i < n; i++ {
		for j := n - i - 1; j < n; j++ {
			fmt.Print(n)
		}
		fmt.Print("\n")
	}
}


//star table 1-9 count input 1:* 2: 3:** ...
package main
import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    var N int
    scanner.Scan()
    fmt.Sscan(scanner.Text(),&N)
    var t [10]int    
    scanner.Scan()
    inputs := strings.Split(scanner.Text()," ")
    for i := 0; i < N; i++ {
        sampleValue,_ := strconv.ParseInt(inputs[i],10,32)
        t[sampleValue]++
    }    
    for i := 1; i < 10; i++ {
        fmt.Printf("%v:",i)
        for x := 0; x< t[i]; x++ {
            fmt.Print("*")
        }
        fmt.Print("\n")
    }    
}

//Split integer at position and apply operation
package main
import "fmt"
import "strconv"
func main() {
    var O string
    var X, N int
    fmt.Scan(&O, &X, &N)    
    S := fmt.Sprint(N)    
    a,_ := strconv.ParseInt(S[:X],10,32)
    b,_ := strconv.ParseInt(S[X:],10,32)    
    switch O {
        case "+":
        fmt.Println(a+b)
        case "-":
        fmt.Println(a-b)
        case "*":
        fmt.Println(a*b)
        case "/":
        fmt.Println(a/b)
    }
}

//integer string processing replace n by 9-n if lower
package main
import "fmt"
func main() {
    var N string
    fmt.Scan(&N)    
    var out string    
    for i := 0; i < len(N); i++ {
        x := N[i] - '0'
        if 9-x < x {
            out += fmt.Sprint(9-x)
        } else {
            out += fmt.Sprint(x)
        }
    }
    fmt.Print(out)
}

//word starts with Capital H !!!!!!! WARNING GOT ONLY 62% at this one
package main
import "fmt"
import "os"
import "bufio"
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    word := scanner.Text()    
    fmt.Println(word[0] == 'H')// Write answer to stdout
}

//linear function a*x+b
package main
import "fmt"
func main() {
    var a, b int
    fmt.Scan(&a, &b)    
    var n int
    fmt.Scan(&n)    
    for i := 0; i < n; i++ {
        var x int
        fmt.Scan(&x)
        fmt.Println(a*x+b)
    }
}

//Print reverse words
package main
import "fmt"
import "os"
import "bufio"
import "strings"
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    S := scanner.Text()    
    words := strings.Split(S," ")    
    for n,word := range words {
        for i:= len(word)-1; i>=0; i-- {
            fmt.Printf("%c",word[i])
        }
        if n != len(words)-1 {
            fmt.Print(" ")
        }
    }
}

//Lowest of two numbers
package main
import "fmt"
func main() {
    var N int
    fmt.Scan(&N)    
    var M int
    fmt.Scan(&M)
    if N<M {
    	fmt.Println(N)
    } else {
    	fmt.Println(M)
	}
}

//binary representation is only composed of 1s
package main
import "fmt"
func main() {
    var n int
    fmt.Scan(&n)    
    a := 2147483648    
    oneFound := false    
    for i:=0; i < 32; i++ {
        if n&a > 0 {
            oneFound = true
        } else {
            if oneFound {
                fmt.Println(false)// Write answer to stdout
                return
            }
        }
        a>>=1
    }
    fmt.Println(true)// Write answer to stdout
}

//n*n*100
package main
import "fmt"
func main() {
    var n int
    fmt.Scan(&n)
    fmt.Println(n*n*100)// Write answer to stdout
}


//Binary NOT
package main
import "fmt"
func main() {
    var B string
    fmt.Scan(&B)    
    for i:=0; i < len(B); i++ {
        if B[i] == '0' {
            fmt.Print(1)
        } else {            
            fmt.Print(0)
        }
    }
}

//word rotations !!!!!!!!!!! WARNING ONLY GOT 80%
package main
import "fmt"
import "os"
import "bufio"
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    word := scanner.Text()
    for i:=0; i <= len(word); i++ {
        fmt.Println(word)
        word = word[len(word)-1:] + word[:len(word)-1]
    }
}


//DNA Complementary
package main
import "fmt"
func main() {
    var DNA string
    fmt.Scan(&DNA)
    for i:=0; i < len(DNA); i++ {
        switch DNA[i] {
            case 'A':
            fmt.Print("T")
            case 'T':
            fmt.Print("A")
            case 'C':
            fmt.Print("G")
            case 'G':
            fmt.Print("C")            
        }
    }
}

//6 rebounds !!!!!! This one was shortest code
package main
import "fmt"
func main() {
    var H int
    fmt.Scan(&H)
    fmt.Println(H*2*2*2*2*2*2)
}

//Cut word in two and print last part before first part + handle case where N < len(word)
package main
import "fmt"
import "os"
import "bufio"
func main() {
    scanner := bufio.NewScanner(os.Stdin)
    var N int
    scanner.Scan()
    fmt.Sscan(scanner.Text(),&N)    
    scanner.Scan()
    word := scanner.Text()
    M:=N%len(word)    
    fmt.Println(word[M:] + word[:M])
}