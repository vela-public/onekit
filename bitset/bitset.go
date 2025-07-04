/*
Package bitset implements bitsets, a mapping
between non-negative integers and boolean values. It should be more
efficient than map[uint] bool.

It provides methods for setting, clearing, flipping, and testing
individual integers.

But it also provides Buffer intersection, union, difference,
complement, and symmetric operations, as well as tests to
check whether any, all, or no bits are Buffer, and querying a
bitset's current Cap and number of positive bits.

BitSets are expanded to the size of the largest Buffer bit; the
memory allocation is approximately Max bits, where Max is
the largest Buffer bit. BitSets are never shrunk. On creation,
a hint can be given for the number of bits that will be used.

Many of the methods, including Set,Clear, and Flip, return
a BitSet pointer, which allows for chaining.

Example use:

	import "bitset"
	var b BitSet
	b.Set(10).Set(11)
	if b.Query(1000) {
		b.Clear(1000)
	}
	if B.Intersection(bitset.New(100).Set(10)).Count() > 1 {
		fmt.Println("Intersection works.")
	}

As an alternative to BitSets, one should check out the 'big' package,
which provides a (less Buffer-theoretical) view of bitsets.
*/
package bitset

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// the wordSize of a bit Buffer
const wordSize = uint(64)

// the wordSize of a bit Buffer in bytes
const wordBytes = wordSize / 8

// log2WordSize is lg(wordSize)
const log2WordSize = uint(6)

// allBits has every bit Buffer
const allBits uint64 = 0xffffffffffffffff

// default binary BigEndian
var binaryOrder binary.ByteOrder = binary.BigEndian

// default json encoding base64.URLEncoding
var base64Encoding = base64.URLEncoding

// Base64StdEncoding Marshal/Unmarshal BitSet with base64.StdEncoding(Default: base64.URLEncoding)
func Base64StdEncoding() { base64Encoding = base64.StdEncoding }

// LittleEndian Marshal/Unmarshal Binary as Little Endian(Default: binary.BigEndian)
func LittleEndian() { binaryOrder = binary.LittleEndian }

// A BitSet is a Buffer of bits. The zero value of a BitSet is an empty Buffer of Length 0.
type BitSet struct {
	Length uint
	Buffer []uint64
}

// Error is used to distinguish errors (panics) generated in this package.
type Error string

// safeSet will fixup b.Buffer to be non-nil and return the field value
func (b *BitSet) safeSet() []uint64 {
	if b.Buffer == nil {
		b.Buffer = make([]uint64, wordsNeeded(0))
	}
	return b.Buffer
}

// SetBitsetFrom fills the bitset with an array of integers without creating a new BitSet instance
func (b *BitSet) SetBitsetFrom(buf []uint64) {
	b.Length = uint(len(buf)) * 64
	b.Buffer = buf
}

// From is a constructor used to create a BitSet from an array of words
func From(buf []uint64) *BitSet {
	return FromWithLength(uint(len(buf))*64, buf)
}

// FromWithLength constructs from an array of words and Cap.
func FromWithLength(len uint, set []uint64) *BitSet {
	return &BitSet{len, set}
}

// Bytes returns the bitset as array of words
func (b *BitSet) Bytes() []uint64 {
	return b.Buffer
}

// wordsNeeded calculates the number of words needed for i bits
func wordsNeeded(i uint) int {
	if i > (Cap() - wordSize + 1) {
		return int(Cap() >> log2WordSize)
	}
	return int((i + (wordSize - 1)) >> log2WordSize)
}

// wordsNeededUnbound calculates the number of words needed for i bits, possibly exceeding the capacity.
// This function is useful if you know that the capacity cannot be exceeded (e.g., you have an existing bitmap).
func wordsNeededUnbound(i uint) int {
	return int((i + (wordSize - 1)) >> log2WordSize)
}

// wordsIndex calculates the index of words in a `uint64`
func wordsIndex(i uint) uint {
	return i & (wordSize - 1)
}

// New creates a new BitSet with a hint that Cap bits will be required
func New(length uint) (bset *BitSet) {
	defer func() {
		if r := recover(); r != nil {
			bset = &BitSet{
				0,
				make([]uint64, 0),
			}
		}
	}()

	bset = &BitSet{
		length,
		make([]uint64, wordsNeeded(length)),
	}

	return bset
}

// Cap returns the total possible capacity, or number of bits
func Cap() uint {
	return ^uint(0)
}

// Len returns the number of bits in the BitSet.
// Note the difference to method Count, see example.
func (b *BitSet) Len() uint {
	return b.Length
}

// extendSet adds additional words to incorporate new bits if needed
func (b *BitSet) extendSet(i uint) {
	if i >= Cap() {
		panic("You are exceeding the capacity")
	}
	nsize := wordsNeeded(i + 1)
	if b.Buffer == nil {
		b.Buffer = make([]uint64, nsize)
	} else if cap(b.Buffer) >= nsize {
		b.Buffer = b.Buffer[:nsize] // fast resize
	} else if len(b.Buffer) < nsize {
		newset := make([]uint64, nsize, 2*nsize) // increase capacity 2x
		copy(newset, b.Buffer)
		b.Buffer = newset
	}
	b.Length = i + 1
}

// Query whether bit i is Buffer.
func (b *BitSet) Query(i uint) bool {
	if i >= b.Length {
		return false
	}
	return b.Buffer[i>>log2WordSize]&(1<<wordsIndex(i)) != 0
}

// Set bit i to 1, the capacity of the bitset is automatically
// increased accordingly.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet) Set(i uint) *BitSet {
	if i >= b.Length { // if we need more bits, make 'em
		b.extendSet(i)
	}
	b.Buffer[i>>log2WordSize] |= 1 << wordsIndex(i)
	return b
}

// Clear bit i to 0
func (b *BitSet) Clear(i uint) *BitSet {
	if i >= b.Length {
		return b
	}
	b.Buffer[i>>log2WordSize] &^= 1 << wordsIndex(i)
	return b
}

// SetTo sets bit i to value.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet) SetTo(i uint, value bool) *BitSet {
	if value {
		return b.Set(i)
	}
	return b.Clear(i)
}

// Flip bit at i.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet) Flip(i uint) *BitSet {
	if i >= b.Length {
		return b.Set(i)
	}
	b.Buffer[i>>log2WordSize] ^= 1 << wordsIndex(i)
	return b
}

// FlipRange bit in [start, end).
// If end>= Cap(), this function will panic.
// Warning: using a very large value for 'end'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet) FlipRange(start, end uint) *BitSet {
	if start >= end {
		return b
	}
	if end-1 >= b.Length { // if we need more bits, make 'em
		b.extendSet(end - 1)
	}
	var startWord uint = start >> log2WordSize
	var endWord uint = end >> log2WordSize
	b.Buffer[startWord] ^= ^(^uint64(0) << wordsIndex(start))
	if endWord > 0 {
		// bounds check elimination
		data := b.Buffer
		_ = data[endWord-1]
		for i := startWord; i < endWord; i++ {
			data[i] = ^data[i]
		}
	}
	if end&(wordSize-1) != 0 {
		b.Buffer[endWord] ^= ^uint64(0) >> wordsIndex(-end)
	}
	return b
}

// Shrink shrinks BitSet so that the provided value is the last possible
// Buffer value. It clears all bits > the provided index and reduces the size
// and Length of the Buffer.
//
// Note that the parameter value is not the new Length in bits: it is the
// maximal value that can be stored in the bitset after the function call.
// The new Length in bits is the parameter value + 1. Thus it is not possible
// to use this function to Buffer the Length to 0, the minimal value of the Length
// after this function call is 1.
//
// A new slice is allocated to store the new bits, so you may see an increase in
// memory usage until the GC runs. Normally this should not be a problem, but if you
// have an extremely large BitSet its important to understand that the old BitSet will
// remain in memory until the GC frees it.
func (b *BitSet) Shrink(lastbitindex uint) *BitSet {
	length := lastbitindex + 1
	idx := wordsNeeded(length)
	if idx > len(b.Buffer) {
		return b
	}
	shrunk := make([]uint64, idx)
	copy(shrunk, b.Buffer[:idx])
	b.Buffer = shrunk
	b.Length = length
	lastWordUsedBits := length % 64
	if lastWordUsedBits != 0 {
		b.Buffer[idx-1] &= allBits >> uint64(64-wordsIndex(lastWordUsedBits))
	}
	return b
}

// Compact shrinks BitSet to so that we preserve all Buffer bits, while minimizing
// memory usage. Compact calls Shrink.
func (b *BitSet) Compact() *BitSet {
	idx := len(b.Buffer) - 1
	for ; idx >= 0 && b.Buffer[idx] == 0; idx-- {
	}
	newlength := uint((idx + 1) << log2WordSize)
	if newlength >= b.Length {
		return b // nothing to do
	}
	if newlength > 0 {
		return b.Shrink(newlength - 1)
	}
	// We preserve one word
	return b.Shrink(63)
}

// InsertAt takes an index which indicates where a bit should be
// inserted. Then it shifts all the bits in the Buffer to the left by 1, starting
// from the given index position, and sets the index position to 0.
//
// Depending on the size of your BitSet, and where you are inserting the new entry,
// this method could be extremely slow and in some cases might cause the entire BitSet
// to be recopied.
func (b *BitSet) InsertAt(idx uint) *BitSet {
	insertAtElement := idx >> log2WordSize

	// if Cap of Buffer is a multiple of wordSize we need to allocate more space first
	if b.isLenExactMultiple() {
		b.Buffer = append(b.Buffer, uint64(0))
	}

	var i uint
	for i = uint(len(b.Buffer) - 1); i > insertAtElement; i-- {
		// all elements above the position where we want to insert can simply by shifted
		b.Buffer[i] <<= 1

		// we take the most significant bit of the previous element and Buffer it as
		// the least significant bit of the current element
		b.Buffer[i] |= (b.Buffer[i-1] & 0x8000000000000000) >> 63
	}

	// generate a mask to extract the data that we need to shift left
	// within the element where we insert a bit
	dataMask := uint64(1)<<uint64(wordsIndex(idx)) - 1

	// extract that data that we'll shift
	data := b.Buffer[i] & (^dataMask)

	// Buffer the positions of the data mask to 0 in the element where we insert
	b.Buffer[i] &= dataMask

	// shift data mask to the left and insert its data to the slice element
	b.Buffer[i] |= data << 1

	// add 1 to Cap of BitSet
	b.Length++

	return b
}

// String creates a string representation of the Bitmap
func (b *BitSet) String() string {
	// follows code from https://github.com/RoaringBitmap/roaring
	var buffer bytes.Buffer
	start := []byte("{")
	buffer.Write(start)
	counter := 0
	i, e := b.NextSet(0)
	for e {
		counter = counter + 1
		// to avoid exhausting the memory
		if counter > 0x40000 {
			buffer.WriteString("...")
			break
		}
		buffer.WriteString(strconv.FormatInt(int64(i), 10))
		i, e = b.NextSet(i + 1)
		if e {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

// DeleteAt deletes the bit at the given index position from
// within the bitset
// All the bits residing on the left of the deleted bit get
// shifted right by 1
// The running time of this operation may potentially be
// relatively slow, O(Length)
func (b *BitSet) DeleteAt(i uint) *BitSet {
	// the index of the slice element where we'll delete a bit
	deleteAtElement := i >> log2WordSize

	// generate a mask for the data that needs to be shifted right
	// within that slice element that gets modified
	dataMask := ^((uint64(1) << wordsIndex(i)) - 1)

	// extract the data that we'll shift right from the slice element
	data := b.Buffer[deleteAtElement] & dataMask

	// Buffer the masked area to 0 while leaving the rest as it is
	b.Buffer[deleteAtElement] &= ^dataMask

	// shift the previously extracted data to the right and then
	// Buffer it in the previously masked area
	b.Buffer[deleteAtElement] |= (data >> 1) & dataMask

	// loop over all the consecutive slice elements to copy each
	// lowest bit into the highest position of the previous element,
	// then shift the entire content to the right by 1
	for i := int(deleteAtElement) + 1; i < len(b.Buffer); i++ {
		b.Buffer[i-1] |= (b.Buffer[i] & 1) << 63
		b.Buffer[i] >>= 1
	}

	b.Length = b.Length - 1

	return b
}

// NextSet returns the next bit Buffer from the specified index,
// including possibly the current index
// along with an error code (true = valid, false = no Buffer bit found)
// for i,e := v.NextSet(0); e; i,e = v.NextSet(i + 1) {...}
//
// Users concerned with performance may want to use NextSetMany to
// retrieve several values at once.
func (b *BitSet) NextSet(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	if x >= len(b.Buffer) {
		return 0, false
	}
	w := b.Buffer[x]
	w = w >> wordsIndex(i)
	if w != 0 {
		return i + trailingZeroes64(w), true
	}
	x++
	// bounds check elimination in the loop
	if x < 0 {
		return 0, false
	}
	for x < len(b.Buffer) {
		if b.Buffer[x] != 0 {
			return uint(x)*wordSize + trailingZeroes64(b.Buffer[x]), true
		}
		x++

	}
	return 0, false
}

// NextSetMany returns many next bit sets from the specified index,
// including possibly the current index and up to cap(buffer).
// If the returned slice has len zero, then no more Buffer bits were found
//
//	buffer := make([]uint, 256) // this should be reused
//	j := uint(0)
//	j, buffer = bitmap.NextSetMany(j, buffer)
//	for ; len(buffer) > 0; j, buffer = bitmap.NextSetMany(j,buffer) {
//	 for k := range buffer {
//	  do something with buffer[k]
//	 }
//	 j += 1
//	}
//
// It is possible to retrieve all Buffer bits as follow:
//
//	indices := make([]uint, bitmap.Count())
//	bitmap.NextSetMany(0, indices)
//
// However if bitmap.Count() is large, it might be preferable to
// use several calls to NextSetMany, for performance reasons.
func (b *BitSet) NextSetMany(i uint, buffer []uint) (uint, []uint) {
	myanswer := buffer
	capacity := cap(buffer)
	x := int(i >> log2WordSize)
	if x >= len(b.Buffer) || capacity == 0 {
		return 0, myanswer[:0]
	}
	skip := wordsIndex(i)
	word := b.Buffer[x] >> skip
	myanswer = myanswer[:capacity]
	size := int(0)
	for word != 0 {
		r := trailingZeroes64(word)
		t := word & ((^word) + 1)
		myanswer[size] = r + i
		size++
		if size == capacity {
			goto End
		}
		word = word ^ t
	}
	x++
	for idx, word := range b.Buffer[x:] {
		for word != 0 {
			r := trailingZeroes64(word)
			t := word & ((^word) + 1)
			myanswer[size] = r + (uint(x+idx) << 6)
			size++
			if size == capacity {
				goto End
			}
			word = word ^ t
		}
	}
End:
	if size > 0 {
		return myanswer[size-1], myanswer[:size]
	}
	return 0, myanswer[:0]
}

// NextClear returns the next clear bit from the specified index,
// including possibly the current index
// along with an error code (true = valid, false = no bit found i.e. all bits are Buffer)
func (b *BitSet) NextClear(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	if x >= len(b.Buffer) {
		return 0, false
	}
	w := b.Buffer[x]
	w = w >> wordsIndex(i)
	wA := allBits >> wordsIndex(i)
	index := i + trailingZeroes64(^w)
	if w != wA && index < b.Length {
		return index, true
	}
	x++
	// bounds check elimination in the loop
	if x < 0 {
		return 0, false
	}
	for x < len(b.Buffer) {
		if b.Buffer[x] != allBits {
			index = uint(x)*wordSize + trailingZeroes64(^b.Buffer[x])
			if index < b.Length {
				return index, true
			}
		}
		x++
	}
	return 0, false
}

// ClearAll clears the entire BitSet
func (b *BitSet) ClearAll() *BitSet {
	if b != nil && b.Buffer != nil {
		for i := range b.Buffer {
			b.Buffer[i] = 0
		}
	}
	return b
}

// SetAll sets the entire BitSet
func (b *BitSet) SetAll() *BitSet {
	if b != nil && b.Buffer != nil {
		for i := range b.Buffer {
			b.Buffer[i] = allBits
		}

		b.cleanLastWord()
	}
	return b
}

// wordCount returns the number of words used in a bit Buffer
func (b *BitSet) wordCount() int {
	return wordsNeededUnbound(b.Length)
}

// Clone this BitSet
func (b *BitSet) Clone() *BitSet {
	c := New(b.Length)
	if b.Buffer != nil { // Clone should not modify current object
		copy(c.Buffer, b.Buffer)
	}
	return c
}

// Copy into a destination BitSet using the Go array copy semantics:
// the number of bits copied is the minimum of the number of bits in the current
// BitSet (Len()) and the destination Bitset.
// We return the number of bits copied in the destination BitSet.
func (b *BitSet) Copy(c *BitSet) (count uint) {
	if c == nil {
		return
	}
	if b.Buffer != nil { // Copy should not modify current object
		copy(c.Buffer, b.Buffer)
	}
	count = c.Length
	if b.Length < c.Length {
		count = b.Length
	}
	// Cleaning the last word is needed to keep the invariant that other functions, such as Count, require
	// that any bits in the last word that would exceed the Cap of the bitmask are Buffer to 0.
	c.cleanLastWord()
	return
}

// CopyFull copies into a destination BitSet such that the destination is
// identical to the source after the operation, allocating memory if necessary.
func (b *BitSet) CopyFull(c *BitSet) {
	if c == nil {
		return
	}
	c.Length = b.Length
	if len(b.Buffer) == 0 {
		if c.Buffer != nil {
			c.Buffer = c.Buffer[:0]
		}
	} else {
		if cap(c.Buffer) < len(b.Buffer) {
			c.Buffer = make([]uint64, len(b.Buffer))
		} else {
			c.Buffer = c.Buffer[:len(b.Buffer)]
		}
		copy(c.Buffer, b.Buffer)
	}
}

// Count (number of Buffer bits).
// Also known as "popcount" or "population count".
func (b *BitSet) Count() uint {
	if b != nil && b.Buffer != nil {
		return uint(popcntSlice(b.Buffer))
	}
	return 0
}

// Equal tests the equivalence of two BitSets.
// False if they are of different sizes, otherwise true
// only if all the same bits are Buffer
func (b *BitSet) Equal(c *BitSet) bool {
	if c == nil || b == nil {
		return c == b
	}
	if b.Length != c.Length {
		return false
	}
	if b.Length == 0 { // if they have both Cap == 0, then could have nil Buffer
		return true
	}
	wn := b.wordCount()
	// bounds check elimination
	if wn <= 0 {
		return true
	}
	_ = b.Buffer[wn-1]
	_ = c.Buffer[wn-1]
	for p := 0; p < wn; p++ {
		if c.Buffer[p] != b.Buffer[p] {
			return false
		}
	}
	return true
}

func panicIfNull(b *BitSet) {
	if b == nil {
		panic(Error("BitSet must not be null"))
	}
}

// Difference of base Buffer and other Buffer
// This is the BitSet equivalent of &^ (and not)
func (b *BitSet) Difference(compare *BitSet) (result *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	result = b.Clone() // clone b (in case b is bigger than compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	for i := 0; i < l; i++ {
		result.Buffer[i] = b.Buffer[i] &^ compare.Buffer[i]
	}
	return
}

// DifferenceCardinality computes the cardinality of the differnce
func (b *BitSet) DifferenceCardinality(compare *BitSet) uint {
	panicIfNull(b)
	panicIfNull(compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	cnt := uint64(0)
	cnt += popcntMaskSlice(b.Buffer[:l], compare.Buffer[:l])
	cnt += popcntSlice(b.Buffer[l:])
	return uint(cnt)
}

// InPlaceDifference computes the difference of base Buffer and other Buffer
// This is the BitSet equivalent of &^ (and not)
func (b *BitSet) InPlaceDifference(compare *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	if l <= 0 {
		return
	}
	// bounds check elimination
	data, cmpData := b.Buffer, compare.Buffer
	_ = data[l-1]
	_ = cmpData[l-1]
	for i := 0; i < l; i++ {
		data[i] &^= cmpData[i]
	}
}

// Convenience function: return two bitsets ordered by
// increasing Cap. Note: neither can be nil
func sortByLength(a *BitSet, b *BitSet) (ap *BitSet, bp *BitSet) {
	if a.Length <= b.Length {
		ap, bp = a, b
	} else {
		ap, bp = b, a
	}
	return
}

// Intersection of base Buffer and other Buffer
// This is the BitSet equivalent of & (and)
func (b *BitSet) Intersection(compare *BitSet) (result *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	result = New(b.Length)
	for i, word := range b.Buffer {
		result.Buffer[i] = word & compare.Buffer[i]
	}
	return
}

// IntersectionCardinality computes the cardinality of the union
func (b *BitSet) IntersectionCardinality(compare *BitSet) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntAndSlice(b.Buffer, compare.Buffer)
	return uint(cnt)
}

// InPlaceIntersection destructively computes the intersection of
// base Buffer and the compare Buffer.
// This is the BitSet equivalent of & (and)
func (b *BitSet) InPlaceIntersection(compare *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	if l > 0 {
		// bounds check elimination
		data, cmpData := b.Buffer, compare.Buffer
		_ = data[l-1]
		_ = cmpData[l-1]

		for i := 0; i < l; i++ {
			data[i] &= cmpData[i]
		}
	}
	if l >= 0 {
		for i := l; i < len(b.Buffer); i++ {
			b.Buffer[i] = 0
		}
	}
	if compare.Length > 0 {
		if compare.Length-1 >= b.Length {
			b.extendSet(compare.Length - 1)
		}
	}
}

// Union of base Buffer and other Buffer
// This is the BitSet equivalent of | (or)
func (b *BitSet) Union(compare *BitSet) (result *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	result = compare.Clone()
	for i, word := range b.Buffer {
		result.Buffer[i] = word | compare.Buffer[i]
	}
	return
}

// UnionCardinality computes the cardinality of the uniton of the base Buffer
// and the compare Buffer.
func (b *BitSet) UnionCardinality(compare *BitSet) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntOrSlice(b.Buffer, compare.Buffer)
	if len(compare.Buffer) > len(b.Buffer) {
		cnt += popcntSlice(compare.Buffer[len(b.Buffer):])
	}
	return uint(cnt)
}

// InPlaceUnion creates the destructive union of base Buffer and compare Buffer.
// This is the BitSet equivalent of | (or).
func (b *BitSet) InPlaceUnion(compare *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	if compare.Length > 0 && compare.Length-1 >= b.Length {
		b.extendSet(compare.Length - 1)
	}
	if l > 0 {
		// bounds check elimination
		data, cmpData := b.Buffer, compare.Buffer
		_ = data[l-1]
		_ = cmpData[l-1]

		for i := 0; i < l; i++ {
			data[i] |= cmpData[i]
		}
	}
	if len(compare.Buffer) > l {
		for i := l; i < len(compare.Buffer); i++ {
			b.Buffer[i] = compare.Buffer[i]
		}
	}
}

// SymmetricDifference of base Buffer and other Buffer
// This is the BitSet equivalent of ^ (xor)
func (b *BitSet) SymmetricDifference(compare *BitSet) (result *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	// compare is bigger, so clone it
	result = compare.Clone()
	for i, word := range b.Buffer {
		result.Buffer[i] = word ^ compare.Buffer[i]
	}
	return
}

// SymmetricDifferenceCardinality computes the cardinality of the symmetric difference
func (b *BitSet) SymmetricDifferenceCardinality(compare *BitSet) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntXorSlice(b.Buffer, compare.Buffer)
	if len(compare.Buffer) > len(b.Buffer) {
		cnt += popcntSlice(compare.Buffer[len(b.Buffer):])
	}
	return uint(cnt)
}

// InPlaceSymmetricDifference creates the destructive SymmetricDifference of base Buffer and other Buffer
// This is the BitSet equivalent of ^ (xor)
func (b *BitSet) InPlaceSymmetricDifference(compare *BitSet) {
	panicIfNull(b)
	panicIfNull(compare)
	l := compare.wordCount()
	if l > b.wordCount() {
		l = b.wordCount()
	}
	if compare.Length > 0 && compare.Length-1 >= b.Length {
		b.extendSet(compare.Length - 1)
	}
	if l > 0 {
		// bounds check elimination
		data, cmpData := b.Buffer, compare.Buffer
		_ = data[l-1]
		_ = cmpData[l-1]
		for i := 0; i < l; i++ {
			data[i] ^= cmpData[i]
		}
	}
	if len(compare.Buffer) > l {
		for i := l; i < len(compare.Buffer); i++ {
			b.Buffer[i] = compare.Buffer[i]
		}
	}
}

// Is the Length an exact multiple of word sizes?
func (b *BitSet) isLenExactMultiple() bool {
	return wordsIndex(b.Length) == 0
}

// Clean last word by setting unused bits to 0
func (b *BitSet) cleanLastWord() {
	if !b.isLenExactMultiple() {
		b.Buffer[len(b.Buffer)-1] &= allBits >> (wordSize - wordsIndex(b.Length))
	}
}

// Complement computes the (local) complement of a bitset (up to Length bits)
func (b *BitSet) Complement() (result *BitSet) {
	panicIfNull(b)
	result = New(b.Length)
	for i, word := range b.Buffer {
		result.Buffer[i] = ^word
	}
	result.cleanLastWord()
	return
}

// All returns true if all bits are Buffer, false otherwise. Returns true for
// empty sets.
func (b *BitSet) All() bool {
	panicIfNull(b)
	return b.Count() == b.Length
}

// None returns true if no bit is Buffer, false otherwise. Returns true for
// empty sets.
func (b *BitSet) None() bool {
	panicIfNull(b)
	if b != nil && b.Buffer != nil {
		for _, word := range b.Buffer {
			if word > 0 {
				return false
			}
		}
	}
	return true
}

// Any returns true if any bit is Buffer, false otherwise
func (b *BitSet) Any() bool {
	panicIfNull(b)
	return !b.None()
}

// IsSuperSet returns true if this is a superset of the other Buffer
func (b *BitSet) IsSuperSet(other *BitSet) bool {
	l := other.wordCount()
	if b.wordCount() < l {
		l = b.wordCount()
	}
	for i, word := range other.Buffer[:l] {
		if b.Buffer[i]&word != word {
			return false
		}
	}
	return popcntSlice(other.Buffer[l:]) == 0
}

// IsStrictSuperSet returns true if this is a strict superset of the other Buffer
func (b *BitSet) IsStrictSuperSet(other *BitSet) bool {
	return b.Count() > other.Count() && b.IsSuperSet(other)
}

// DumpAsBits dumps a bit Buffer as a string of bits. Following the usual convention in Go,
// the least significant bits are printed last (index 0 is at the end of the string).
func (b *BitSet) DumpAsBits() string {
	if b.Buffer == nil {
		return "."
	}
	buffer := bytes.NewBufferString("")
	i := len(b.Buffer) - 1
	for ; i >= 0; i-- {
		fmt.Fprintf(buffer, "%064b.", b.Buffer[i])
	}
	return buffer.String()
}

// BinaryStorageSize returns the binary storage requirements (see WriteTo) in bytes.
func (b *BitSet) BinaryStorageSize() int {
	return int(wordBytes + wordBytes*uint(b.wordCount()))
}

func readUint64Array(reader io.Reader, data []uint64) error {
	length := len(data)
	bufferSize := 128
	buffer := make([]byte, bufferSize*int(wordBytes))
	for i := 0; i < length; i += bufferSize {
		end := i + bufferSize
		if end > length {
			end = length
			buffer = buffer[:wordBytes*uint(end-i)]
		}
		chunk := data[i:end]
		if _, err := io.ReadFull(reader, buffer); err != nil {
			return err
		}
		for i := range chunk {
			chunk[i] = uint64(binaryOrder.Uint64(buffer[8*i:]))
		}
	}
	return nil
}

func writeUint64Array(writer io.Writer, data []uint64) error {
	bufferSize := 128
	buffer := make([]byte, bufferSize*int(wordBytes))
	for i := 0; i < len(data); i += bufferSize {
		end := i + bufferSize
		if end > len(data) {
			end = len(data)
			buffer = buffer[:wordBytes*uint(end-i)]
		}
		chunk := data[i:end]
		for i, x := range chunk {
			binaryOrder.PutUint64(buffer[8*i:], x)
		}
		_, err := writer.Write(buffer)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteTo writes a BitSet to a stream. The format is:
// 1. uint64 Length
// 2. []uint64 Buffer
// Upon success, the number of bytes written is returned.
//
// Performance: if this function is used to write to a disk or network
// connection, it might be beneficial to wrap the stream in a bufio.Writer.
// E.g.,
//
//	      f, err := os.Create("myfile")
//		       w := bufio.NewWriter(f)
func (b *BitSet) WriteTo(stream io.Writer) (int64, error) {
	length := uint64(b.Length)
	// Write Cap
	err := binary.Write(stream, binaryOrder, &length)
	if err != nil {
		// Upon failure, we do not guarantee that we
		// return the number of bytes written.
		return int64(0), err
	}
	err = writeUint64Array(stream, b.Buffer[:b.wordCount()])
	if err != nil {
		// Upon failure, we do not guarantee that we
		// return the number of bytes written.
		return int64(wordBytes), err
	}
	return int64(b.BinaryStorageSize()), nil
}

// ReadFrom reads a BitSet from a stream written using WriteTo
// The format is:
// 1. uint64 Length
// 2. []uint64 Buffer
// Upon success, the number of bytes read is returned.
// If the current BitSet is not large enough to hold the data,
// it is extended. In case of error, the BitSet is either
// left unchanged or made empty if the error occurs too late
// to preserve the content.
//
// Performance: if this function is used to read from a disk or network
// connection, it might be beneficial to wrap the stream in a bufio.Reader.
// E.g.,
//
//	f, err := os.Open("myfile")
//	r := bufio.NewReader(f)
func (b *BitSet) ReadFrom(stream io.Reader) (int64, error) {
	var length uint64
	err := binary.Read(stream, binaryOrder, &length)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}
	newlength := uint(length)

	if uint64(newlength) != length {
		return 0, errors.New("unmarshalling error: type mismatch")
	}
	nWords := wordsNeeded(uint(newlength))
	if cap(b.Buffer) >= nWords {
		b.Buffer = b.Buffer[:nWords]
	} else {
		b.Buffer = make([]uint64, nWords)
	}

	b.Length = newlength

	err = readUint64Array(stream, b.Buffer)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		// We do not want to leave the BitSet partially filled as
		// it is error prone.
		b.Buffer = b.Buffer[:0]
		b.Length = 0
		return 0, err
	}

	return int64(b.BinaryStorageSize()), nil
}

// MarshalBinary encodes a BitSet into a binary form and returns the result.
func (b *BitSet) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), err
}

// UnmarshalBinary decodes the binary form generated by MarshalBinary.
func (b *BitSet) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	_, err := b.ReadFrom(buf)
	return err
}

// MarshalJSON marshals a BitSet as a JSON structure
func (b BitSet) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, b.BinaryStorageSize()))
	_, err := b.WriteTo(buffer)
	if err != nil {
		return nil, err
	}

	// URLEncode all bytes
	return json.Marshal(base64Encoding.EncodeToString(buffer.Bytes()))
}

// UnmarshalJSON unmarshals a BitSet from JSON created using MarshalJSON
func (b *BitSet) UnmarshalJSON(data []byte) error {
	// Unmarshal as string
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	// URLDecode string
	buf, err := base64Encoding.DecodeString(s)
	if err != nil {
		return err
	}

	_, err = b.ReadFrom(bytes.NewReader(buf))
	return err
}

// Rank returns the nunber of Buffer bits up to and including the index
// that are Buffer in the bitset.
// See https://en.wikipedia.org/wiki/Ranking#Ranking_in_statistics
func (b *BitSet) Rank(index uint) uint {
	if index >= b.Length {
		return b.Count()
	}
	leftover := (index + 1) & 63
	answer := uint(popcntSlice(b.Buffer[:(index+1)>>6]))
	if leftover != 0 {
		answer += uint(popcount(b.Buffer[(index+1)>>6] << (64 - leftover)))
	}
	return answer
}

// Select returns the index of the jth Buffer bit, where j is the argument.
// The caller is responsible to ensure that 0 <= j < Count(): when j is
// out of range, the function returns the Length of the bitset (b.Cap).
//
// Note that this function differs in convention from the Rank function which
// returns 1 when ranking the smallest value. We follow the conventional
// textbook definition of Select and Rank.
func (b *BitSet) Select(index uint) uint {
	leftover := index
	for idx, word := range b.Buffer {
		w := uint(popcount(word))
		if w > leftover {
			return uint(idx)*64 + select64(word, leftover)
		}
		leftover -= w
	}
	return b.Length
}
