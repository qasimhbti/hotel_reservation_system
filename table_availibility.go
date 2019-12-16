package main

import (
	"database/sql"
	"fmt"
	"time"
)

// TableStatus represent availibility of tables in reservation
type TableStatus struct {
	TableResStatus []reserveTable
}

type reserveTable struct {
	TableID      int
	BookingDates []bookingDate
}

type bookingDate struct {
	BookDate     string
	BookingSlots []bookingSlot
}

type bookingSlot struct {
	BookingTime  int64
	CheckoutTime int64
}

var tableStatus TableStatus

type bookingDetails struct {
	BDate time.Time
	BTime int64
	CTime int64
}

func (b *bookingDetails) checkTableAvailiblityTimeSlot(tableID int) int {
	for _, resTable := range tableStatus.TableResStatus {
		fmt.Println("resTable.TableID :", resTable.TableID)
		if tableID == resTable.TableID {
			for _, bookingDate := range resTable.BookingDates {
				fmt.Println("bookingDate.BookDate :", bookingDate.BookDate)
				if bookingDate.BookDate == b.BDate.Format(dateLayout) {
					for _, bookingSlot := range bookingDate.BookingSlots {
						fmt.Println("bookingSlot.BookingTime :", bookingSlot.BookingTime)
						fmt.Println("bookingSlot.CheckoutTime :", bookingSlot.CheckoutTime)
						if b.BTime >= bookingSlot.BookingTime && b.BTime <= bookingSlot.CheckoutTime {
							return -1
						} else if b.BTime < bookingSlot.BookingTime && b.CTime > bookingSlot.BookingTime {
							return -1
						}
					}
				}
			}
		}
	}
	return tableID
}

func (b *bookingDetails) updateReserveTableStatus(tableID int) {

	bookSlot := new(bookingSlot)
	bookSlot.BookingTime = b.BTime
	bookSlot.CheckoutTime = b.CTime

	bookDate := new(bookingDate)
	bookDate.BookDate = b.BDate.Format(dateLayout)
	bookDate.BookingSlots = append(bookDate.BookingSlots, *bookSlot)

	resTable := new(reserveTable)
	resTable.TableID = tableID
	resTable.BookingDates = append(resTable.BookingDates, *bookDate)

	tableStatus.TableResStatus = append(tableStatus.TableResStatus, *resTable)
}

func processCheckOut(db *sql.DB, resRes *reservationResult) {
	fmt.Println("-----schedule checkout------")
	fmt.Println("reservation reseult :", resRes)
	fmt.Println("tableStatus :", tableStatus)
	//k := 0
	for _, resTable := range tableStatus.TableResStatus {
		fmt.Println("resTable.TableID :", resTable.TableID)
		if resRes.TableID == resTable.TableID {
			for _, bookingDate := range resTable.BookingDates {
				fmt.Println("bookingDate.BookDate :", bookingDate.BookDate)
				if resRes.BookingDate.Format(dateLayout) == bookingDate.BookDate {
					for k2, bSlot := range bookingDate.BookingSlots {
						if resRes.BookingTime == bSlot.BookingTime {
							bookingDate.BookingSlots[k2] = bookingSlot{}
						}
					}
				}
			}
		}
	}

	fmt.Println("tableStatus :", tableStatus)
}
