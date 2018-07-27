/*
 * Copyright (c) 2002-2018 "Neo4j,"
 * Neo4j Sweden AB [http://neo4j.com]
 *
 * This file is part of Neo4j.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration_tests

import (
	"math/rand"
	"time"

	. "github.com/neo4j/neo4j-go-driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Temporal Types", func() {
	const (
		numberOfRandomValues = 2000
	)

	var (
		err     error
		driver  Driver
		session *Session
		result  *Result
	)

	rand.Seed(time.Now().UnixNano())

	BeforeEach(func() {
		driver, err = NewDriver(singleInstanceUri, BasicAuth(username, password, ""))
		Expect(err).To(BeNil())
		Expect(driver).NotTo(BeNil())

		if VersionOfDriver(driver).LessThan(V3_4_0) {
			Skip("temporal types are only available after neo4j 3.4.0 release")
		}

		session, err = driver.Session(AccessModeWrite)
		Expect(err).To(BeNil())
		Expect(session).NotTo(BeNil())
	})

	AfterEach(func() {
		if session != nil {
			session.Close()
		}

		if driver != nil {
			driver.Close()
		}
	})

	randomDuration := func() Duration {
		sign := int64(1)
		if rand.Intn(2) == 0 {
			sign = -sign
		}

		return DurationOf(
			sign*rand.Int63(),
			sign*rand.Int63(),
			sign*rand.Int63(),
			rand.Intn(1000000000))
	}

	randomLocalDate := func() Date {
		sign := 1
		if rand.Intn(2) == 0 {
			sign = -sign
		}

		return DateOf(
			time.Date(
				sign*rand.Intn(9999),
				time.Month(rand.Intn(12)+1),
				rand.Intn(28)+1,
				0, 0, 0, 0, time.Local))
	}

	randomLocalDateTime := func() LocalDateTime {
		sign := 1
		if rand.Intn(2) == 0 {
			sign = -sign
		}

		return LocalDateTimeOf(
			time.Date(
				sign*rand.Intn(9999),
				time.Month(rand.Intn(12)+1),
				rand.Intn(28)+1,
				rand.Intn(24),
				rand.Intn(60),
				rand.Intn(60),
				rand.Intn(1000000000),
				time.Local))
	}

	randomLocalTime := func() LocalTime {
		return LocalTimeOf(
			time.Date(
				0, 0, 0,
				rand.Intn(24),
				rand.Intn(60),
				rand.Intn(60),
				rand.Intn(1000000000),
				time.Local))
	}

	randomOffsetTime := func() OffsetTime {
		sign := 1
		if rand.Intn(2) == 0 {
			sign = -sign
		}

		return OffsetTimeOf(
			time.Date(
				0, 0, 0,
				rand.Intn(24),
				rand.Intn(60),
				rand.Intn(60),
				rand.Intn(1000000000),
				time.FixedZone("Offset", sign*rand.Intn(64800))))
	}

	randomOffsetDateTime := func() time.Time {
		sign := 1
		if rand.Intn(2) == 0 {
			sign = -sign
		}

		return time.Date(
			rand.Intn(300)+1900,
			time.Month(rand.Intn(12)+1),
			rand.Intn(28)+1,
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1000000000),
			time.FixedZone("Offset", sign*rand.Intn(64800)))
	}

	randomZonedDateTime := func() time.Time {
		var zones = []string{
			"Africa/Harare", "America/Aruba", "Africa/Nairobi", "America/Dawson", "Asia/Beirut", "Asia/Tashkent",
			"Canada/Eastern", "Europe/Malta", "Europe/Volgograd", "Indian/Kerguelen", "Etc/GMT+3",
		}

		location, err := time.LoadLocation(zones[rand.Intn(len(zones))])
		Expect(err).To(BeNil())

		return time.Date(
			rand.Intn(300)+1900,
			time.Month(rand.Intn(12)+1),
			rand.Intn(28)+1,
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1000000000),
			location)
	}

	testReceive := func(query string, expected interface{}) {
		result, err = session.Run(query, nil)
		Expect(err).To(BeNil())

		if result.Next() {
			var received = result.Record().GetByIndex(0)

			Expect(received).To(Equal(expected))
		}
		Expect(result.Err()).To(BeNil())
		Expect(result.Next()).To(BeFalse())
	}

	testSendAndReceive := func(query string, data interface{}, expected []interface{}) {
		result, err = session.Run(query, &map[string]interface{}{"x": data})
		Expect(err).To(BeNil())

		if result.Next() {
			var received = result.Record().Values()

			Expect(received).To(Equal(expected))
		}
		Expect(result.Err()).To(BeNil())
		Expect(result.Next()).To(BeFalse())
	}

	testSendAndReceiveValue := func(value interface{}) {
		result, err = session.Run("CREATE (n:Node {value: $value}) RETURN n.value", &map[string]interface{}{"value": value})
		Expect(err).To(BeNil())

		if result.Next() {
			var received = result.Record().GetByIndex(0)

			Expect(received).To(Equal(value))
		}
		Expect(result.Err()).To(BeNil())
		Expect(result.Next()).To(BeFalse())
	}

	Context("Receive", func() {
		It("duration", func() {
			testReceive("RETURN duration({ months: 16, days: 45, seconds: 120, nanoseconds: 187309812 })", DurationOf(16, 45, 120, 187309812))
		})

		It("date", func() {
			testReceive("RETURN date({ year: 1994, month: 11, day: 15 })", DateOf(time.Date(1994, 11, 15, 0, 0, 0, 0, time.Local)))
		})

		It("local time", func() {
			testReceive("RETURN localtime({ hour: 23, minute: 49, second: 59, nanosecond: 999999999 })", LocalTimeOf(time.Date(0, 0, 0, 23, 49, 59, 999999999, time.Local)))
		})

		It("offset time", func() {
			testReceive("RETURN time({ hour: 23, minute: 49, second: 59, nanosecond: 999999999, timezone:'+03:00' })", OffsetTimeOf(time.Date(0, 0, 0, 23, 49, 59, 999999999, time.FixedZone("Offset", 3*60*60))))
		})

		It("local date time (test location = UTC)", func() {
			testReceive("RETURN localdatetime({ year: 1859, month: 5, day: 31, hour: 23, minute: 49, second: 59, nanosecond: 999999999 })", LocalDateTimeOf(time.Date(1859, 5, 31, 23, 49, 59, 999999999, time.UTC)))
		})

		It("local date time (test location = local)", func() {
			testReceive("RETURN localdatetime({ year: 1859, month: 5, day: 31, hour: 23, minute: 49, second: 59, nanosecond: 999999999 })", LocalDateTimeOf(time.Date(1859, 5, 31, 23, 49, 59, 999999999, time.Local)))
		})

		It("offset date time", func() {
			testReceive("RETURN datetime({ year: 1859, month: 5, day: 31, hour: 23, minute: 49, second: 59, nanosecond: 999999999, timezone:'+02:30' })", time.Date(1859, 5, 31, 23, 49, 59, 999999999, time.FixedZone("Offset", 150*60)))
		})

		It("zoned date time", func() {
			location, err := time.LoadLocation("Europe/London")
			Expect(err).To(BeNil())

			testReceive("RETURN datetime({ year: 1959, month: 5, day: 31, hour: 23, minute: 49, second: 59, nanosecond: 999999999, timezone:'Europe/London' })", time.Date(1959, 5, 31, 23, 49, 59, 999999999, location))
		})
	})

	Context("Send and Receive", func() {
		It("duration", func() {
			data := DurationOf(14, 35, 75, 789012587)

			testSendAndReceive("WITH $x AS x RETURN x, x.months, x.days, x.seconds, x.millisecondsOfSecond, x.microsecondsOfSecond, x.nanosecondsOfSecond",
				data,
				[]interface{}{
					data,
					int64(14),
					int64(35),
					int64(75),
					int64(789),
					int64(789012),
					int64(789012587),
				})
		})

		It("date", func() {
			data := DateOf(time.Date(1976, 6, 13, 0, 0, 0, 0, time.Local))

			testSendAndReceive("WITH $x AS x RETURN x, x.year, x.month, x.day",
				data,
				[]interface{}{
					data,
					int64(1976),
					int64(6),
					int64(13),
				})
		})

		It("local time", func() {
			data := LocalTimeOf(time.Date(0, 0, 0, 12, 34, 56, 789012587, time.Local))

			testSendAndReceive("WITH $x AS x RETURN x, x.hour, x.minute, x.second, x.millisecond, x.microsecond, x.nanosecond",
				data,
				[]interface{}{
					data,
					int64(12),
					int64(34),
					int64(56),
					int64(789),
					int64(789012),
					int64(789012587),
				})
		})

		It("offset time", func() {
			data := OffsetTimeOf(time.Date(0, 0, 0, 12, 34, 56, 789012587, time.FixedZone("Offset", 90*60)))

			testSendAndReceive("WITH $x AS x RETURN x, x.hour, x.minute, x.second, x.millisecond, x.microsecond, x.nanosecond, x.offset",
				data,
				[]interface{}{
					data,
					int64(12),
					int64(34),
					int64(56),
					int64(789),
					int64(789012),
					int64(789012587),
					"+01:30",
				})
		})

		It("local date time", func() {
			data := LocalDateTimeOf(time.Date(1976, 6, 13, 12, 34, 56, 789012587, time.Local))

			testSendAndReceive("WITH $x AS x RETURN x, x.year, x.month, x.day, x.hour, x.minute, x.second, x.millisecond, x.microsecond, x.nanosecond",
				data,
				[]interface{}{
					data,
					int64(1976),
					int64(6),
					int64(13),
					int64(12),
					int64(34),
					int64(56),
					int64(789),
					int64(789012),
					int64(789012587),
				})
		})

		It("offset date time", func() {
			data := time.Date(1976, 6, 13, 12, 34, 56, 789012587, time.FixedZone("Offset", -90*60))

			testSendAndReceive("WITH $x AS x RETURN x, x.year, x.month, x.day, x.hour, x.minute, x.second, x.millisecond, x.microsecond, x.nanosecond, x.offset",
				data,
				[]interface{}{
					data,
					int64(1976),
					int64(6),
					int64(13),
					int64(12),
					int64(34),
					int64(56),
					int64(789),
					int64(789012),
					int64(789012587),
					"-01:30",
				})
		})

		It("zoned date time", func() {
			location, err := time.LoadLocation("US/Pacific")
			Expect(err).To(BeNil())
			data := time.Date(1959, 5, 31, 23, 49, 59, 999999999, location)

			testSendAndReceive("WITH $x AS x RETURN x, x.year, x.month, x.day, x.hour, x.minute, x.second, x.millisecond, x.microsecond, x.nanosecond, x.timezone",
				data,
				[]interface{}{
					data,
					int64(1959),
					int64(5),
					int64(31),
					int64(23),
					int64(49),
					int64(59),
					int64(999),
					int64(999999),
					int64(999999999),
					"US/Pacific",
				})
		})
	})

	Context("Send and receive random", func() {
		It("duration", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomDuration())
			}
		})

		It("date", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomLocalDate())
			}
		})

		It("local time", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomLocalTime())
			}
		})

		It("offset time", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomOffsetTime())
			}
		})

		It("local date time", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomLocalDateTime())
			}
		})

		It("offset date time", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomOffsetDateTime())
			}
		})

		It("zoned date time", func() {
			for i := 0; i < numberOfRandomValues; i++ {
				testSendAndReceiveValue(randomZonedDateTime())
			}
		})
	})

	Context("Send and receive random arrays", func() {
		It("duration", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomDuration()
			}

			testSendAndReceiveValue(list)
		})

		It("date", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomLocalDate()
			}

			testSendAndReceiveValue(list)
		})

		It("local time", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomLocalTime()
			}

			testSendAndReceiveValue(list)
		})

		It("offset time", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomOffsetTime()
			}

			testSendAndReceiveValue(list)
		})

		It("local date time", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomLocalDateTime()
			}

			testSendAndReceiveValue(list)
		})

		It("offset date time", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomOffsetDateTime()
			}

			testSendAndReceiveValue(list)
		})

		It("zoned date time", func() {
			listSize := rand.Intn(1000)
			list := make([]interface{}, listSize)
			for i := 0; i < listSize; i++ {
				list[i] = randomZonedDateTime()
			}

			testSendAndReceiveValue(list)
		})
	})
})
