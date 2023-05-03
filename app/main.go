/*
Copyright 2017 The Knative Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Based on https://github.com/knative/docs/blob/master/serving/samples/autoscale-go/autoscale.go

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"
)

var version = os.Getenv("VERSION")

// Algorithm from https://stackoverflow.com/a/21854246

// Only primes less than or equal to N will be generated
func allPrimes(N int) []int {

	var x, y, n int
	nsqrt := math.Sqrt(float64(N))

	is_prime := make([]bool, N)

	for x = 1; float64(x) <= nsqrt; x++ {
		for y = 1; float64(y) <= nsqrt; y++ {
			n = 4*(x*x) + y*y
			if n <= N && (n%12 == 1 || n%12 == 5) {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x) + y*y
			if n <= N && n%12 == 7 {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x) - y*y
			if x > y && n <= N && n%12 == 11 {
				is_prime[n] = !is_prime[n]
			}
		}
	}

	for n = 5; float64(n) <= nsqrt; n++ {
		if is_prime[n] {
			for y = n * n; y < N; y += n * n {
				is_prime[y] = false
			}
		}
	}

	is_prime[2] = true
	is_prime[3] = true

	primes := make([]int, 0, 1270606)
	for x = 0; x < len(is_prime)-1; x++ {
		if is_prime[x] {
			primes = append(primes, x)
		}
	}

	// primes is now a slice that contains all primes numbers up to N
	return primes
}

func goBloat(mb int, out chan string) {
	go func() {
		b := make([]byte, mb*1024*1024)
		b[0] = 1
		b[len(b)-1] = 1
		out <- fmt.Sprintf("Allocated %v Mb of memory.\n", mb)
	}()
}

func goPrime(max, concurrent int, out chan string) {
	start := time.Now().UnixNano()
	done := make(chan string)
	for i := 0; i < concurrent; i++ {
		go func(i int) {
			p := allPrimes(max)
			if len(p) > 0 {
				done <- fmt.Sprintf("The largest prime less than %v is %v.\n", max, p[len(p)-1])
			} else {
				done <- fmt.Sprintf("There are no primes smaller than %v.\n", max)
			}
		}(i)
	}
	go func() {
		answer := ""
		for i := 0; i < concurrent; i++ {
			answer = <-done
		}
		end := time.Now().UnixNano()
		durationMs := float64(end-start) / 1000000
		out <- answer + fmt.Sprintf("(calculated %v times in %v milliseconds)\n", concurrent, durationMs)
	}()
}

func goSleep(ms int, out chan string) {
	go func() {
		start := time.Now().UnixNano()
		time.Sleep(time.Duration(ms) * time.Millisecond)
		end := time.Now().UnixNano()
		out <- fmt.Sprintf("Slept for %.2f milliseconds.\n", float64(end-start)/1000000)
	}()
}

var (
	sleepMs         = flag.Int("sleep-ms", 500, "milliseconds to sleep")
	primeNth        = flag.Int("prime-nth", 500000, "calculate largest prime less than")
	primeConcurrent = flag.Int("prime-concurrent", 1, "prime calculations to run concurrently")
	bloatMb         = flag.Int("bloat-mb", 2, "mb of memory to consume")
)

func main() {
	flag.Parse()
	out := make(chan string, 3)
	goBloat(*bloatMb, out)
	goPrime(*primeNth, *primeConcurrent, out)
	goSleep(*sleepMs, out)
	for i := 0; i < 3; i++ {
		fmt.Printf(<-out)
	}
}
