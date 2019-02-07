package tnt

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("this test needs tarantool_box")
	}
	require := require.New(t)

	box, err := NewBox("")
	require.NoError(err)
	defer box.Close()

	conn, err := Connect(box.Listen(), nil)
	require.NoError(err)
	defer conn.Close()

	res, err := conn.Execute(&Ping{})
	require.NoError(err)
	assert.Nil(t, res)
}

func TestSelect(t *testing.T) {
	assert := assert.New(t)

	primaryPort, tearDown := setUp(t)
	if t.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d", primaryPort), nil)
	if !assert.NoError(err) {
		return
	}
	defer conn.Close()

	data, err := conn.Execute(&Select{
		Value: PackInt(0),
		Space: 15,
	})
	assert.Nil(data)
	assert.Error(err)
	assert.Equal("Space 15 does not exist", err.Error())
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)

	primaryPort, tearDown := setUp(t)
	if t.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d", primaryPort), nil)
	if !assert.NoError(err) {
		return
	}
	defer conn.Close()

	value1 := uint32(rand.Int31())
	value2 := uint32(rand.Int31())
	value3 := uint32(rand.Int31())
	value4 := uint32(rand.Int31())

	conn.Execute(&Insert{
		Tuple: Tuple{
			PackInt(value1),
			PackInt(value3),
		},
		Space: 1,
	})

	conn.Execute(&Insert{
		Tuple: Tuple{
			PackInt(value1),
			PackInt(value4),
		},
		Space: 1,
	})

	conn.Execute(&Insert{
		Tuple: Tuple{
			PackInt(value2),
			PackInt(value4),
		},
		Space: 1,
	})

	// select 1

	data, err := conn.Execute(&Select{
		Value: PackInt(value1),
		Space: 1,
	})

	assert.NoError(err)
	assert.Equal(
		[]Tuple{
			Tuple{
				PackInt(value1),
				PackInt(value4),
			},
		},
		data,
	)

	// select 2
	data, err = conn.Execute(&Select{
		Value: PackInt(value4),
		Space: 1,
		Index: 1,
	})

	assert.NoError(err)
	assert.Equal(2, len(data))
	assert.Equal(Bytes(PackInt(value4)), data[0][1])
	assert.Equal(Bytes(PackInt(value4)), data[1][1])
}

func TestDefaultSpace(t *testing.T) {
	assert := assert.New(t)

	primaryPort, tearDown := setUp(t)
	if t.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d/1", primaryPort), nil)
	if !assert.NoError(err) {
		return
	}
	defer conn.Close()

	value1 := uint32(rand.Int31())
	value2 := uint32(rand.Int31())

	conn.Execute(&Insert{
		Tuple: Tuple{
			PackInt(value1),
			PackInt(value2),
		},
	})

	data, err := conn.Execute(&Select{
		Value: PackInt(value1),
	})
	assert.NoError(err)
	assert.Equal(1, len(data))
}

func TestDefaultSpace2(t *testing.T) {
	assert := assert.New(t)

	primaryPort, tearDown := setUp(t)
	if t.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d/24", primaryPort), nil)
	if !assert.NoError(err) {
		return
	}
	defer conn.Close()

	data, err := conn.Execute(&Select{
		Value: PackInt(0),
	})
	assert.Nil(data)
	assert.Error(err)
	assert.Equal("Space 24 does not exist", err.Error())
}

func TestDefaultSpace3(t *testing.T) {
	assert := assert.New(t)

	primaryPort, tearDown := setUp(t)
	if t.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d/24", primaryPort), &Options{
		DefaultSpace: 48,
	})
	if !assert.NoError(err) {
		return
	}
	defer conn.Close()

	data, err := conn.Execute(&Select{
		Value: PackInt(0),
	})
	assert.Nil(data)
	assert.Error(err)
	assert.Equal("Space 48 does not exist", err.Error())
}

func BenchmarkSelect(b *testing.B) {
	primaryPort, tearDown := setUp(b)
	if b.Skipped() {
		return
	}
	defer tearDown()

	conn, err := Connect(fmt.Sprintf("127.0.0.1:%d", primaryPort), nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, err := conn.Execute(&Select{
			Value: PackInt(0),
			Space: 10,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
