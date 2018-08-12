package nomof

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
	"testing"
)

func TestFilterBuilder_Equal(t *testing.T) {
	b := NewBuilder()

	b.Equal("Name", "Taro").
		Equal("Name", "Hanako")

	expectedJoin := "('Name' = ? AND 'Name' = ?)"
	if b.JoinAnd() != expectedJoin {
		t.Fatalf("Not match join string: actual:%s exptected:%s", b.JoinAnd(), expectedJoin)
	}

	expectedArgs := []string{"Taro", "Hanako"}
	if b.Arg[0] != expectedArgs[0] || b.Arg[1] != expectedArgs[1] {
		t.Fatalf("Not match args: actual:%v expected:%s", b.Arg, strings.Join(expectedArgs, ","))
	}
}

func TestFilterBuilder_Op(t *testing.T) {
	b := NewBuilder()

	b.Op("Name1", EQ, "Taro1").
		Op("Name2", NE, "Taro2").
		Op("Name3", LT, "Taro3").
		Op("Name4", LE, "Taro4").
		Op("Name5", GT, "Taro5").
		Op("Name6", GE, "Taro6")

	expectedJoin := "('Name1' = ? AND 'Name2' <> ? AND 'Name3' < ? AND 'Name4' <= ? AND 'Name5' > ? AND 'Name6' >= ?)"
	if b.JoinAnd() != expectedJoin {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expectedJoin)
	}

	for i := 1; i <= 6; i++ {
		if b.Arg[i-1] != fmt.Sprintf("Taro%d", i) {
			t.Fatalf("Not match args: actual:%v", b.Arg)
		}
	}
}

func TestFilterBuilder_Between(t *testing.T) {
	b := NewBuilder()

	b.Between("Name", "Taro", "Hanako")

	expected := "('Name' BETWEEN ? AND ?)"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}

	expectedArgs := []string{"Taro", "Hanako"}
	for i := 0; i < 2; i++ {
		if b.Arg[i] != expectedArgs[i] {
			t.Fatalf("Not match args: actual:%v", b.Arg)
		}
	}
}

func TestFilterBuilder_In(t *testing.T) {
	b := NewBuilder()

	b.In("Name", "Taro1", "Taro2", "Taro3")

	expected := "('Name' IN (?,?,?))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}

	for i := 0; i < 3; i++ {
		if b.Arg[i] != fmt.Sprintf("Taro%d", i+1) {
			t.Fatalf("Not match args: actual:%v", b.Arg)
		}
	}
}

func TestFilterBuilder_AttributeExists(t *testing.T) {
	b := NewBuilder()

	b.AttributeExists("Name")

	expected := "(attribute_exists('Name'))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}
}

func TestFilterBuilder_AttributeNotExists(t *testing.T) {
	b := NewBuilder()

	b.AttributeNotExists("Name")

	expected := "(attribute_not_exists('Name'))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}
}

func TestFilterBuilder_AttributeType(t *testing.T) {
	b := NewBuilder()

	b.AttributeType("Name", S)

	expected := "(attribute_type('Name', ?))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}

	if b.Arg[0] != S {
		t.Fatalf("Not match args: actual:%v", b.Arg)
	}
}

func TestFilterBuilder_BeginsWith(t *testing.T) {
	b := NewBuilder()

	b.BeginsWith("Name", "Taro")

	expected := "(begins_with('Name', ?))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}

	if b.Arg[0] != "Taro" {
		t.Fatalf("Not match args: actual:%v", b.Arg)
	}
}

func TestFilterBuilder_Contains(t *testing.T) {
	b := NewBuilder()

	b.Contains("Name", "Taro")

	expected := "(contains('Name', ?))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}

	if b.Arg[0] != "Taro" {
		t.Fatalf("Not match args: actual:%v", b.Arg)
	}
}

func TestFilterBuilder_Size(t *testing.T) {
	b := NewBuilder()

	b.Size("Name")

	expected := "(size('Name'))"
	if b.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b.JoinAnd(), expected)
	}
}

func TestFilterBuilder_JoinAndOr(t *testing.T) {
	b1 := NewBuilder()
	b2 := NewBuilder()

	b1.Equal("Name", "Taro").Equal("Name", "Hanako")
	b2.Equal("Age", 1)

	b2.Append(b1.JoinOr(), b1.Arg)

	expected := "('Age' = ? AND ('Name' = ? OR 'Name' = ?))"
	if b2.JoinAnd() != expected {
		t.Fatalf("Not match join string: actual:%s expected:%s", b2.JoinAnd(), expected)
	}

	expectedArgs := []interface{}{1, "Taro", "Hanako"}
	for i := 0; i < 3; i++ {
		if b2.Arg[i] != expectedArgs[i] {
			t.Fatalf("Not match args: actual:%v", b2.Arg)
		}
	}
}

type Sample struct {
	PK   string `dynamo:"PK,hash"`
	SK   string `dynamo:"SK,range"`
	Name string `dynamo:"Name"`
}

func checkDynamo(t *testing.T) {
	if os.Getenv("DYNAMO_ENDPOINT") == "" {
		t.Skip("Not config dynamodb")
	}
}

func ConnectDB() (*dynamo.Table, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Endpoint:    aws.String(os.Getenv("DYNAMO_ENDPOINT")),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", "dummy"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create session")
	}

	db := dynamo.New(sess)

	db.Table("Samples").DeleteTable().Run()

	err = db.CreateTable("Samples", Sample{}).Run()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create table")
	}

	table := db.Table("Samples")

	if os.Getenv("DEBUG") != "" {
		describeTable(&table)
	}

	return &table, nil
}

func describeTable(table *dynamo.Table) {
	desc, err := table.Describe().Run()
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("%#v", desc)
	}
}

func TestFilterBuilder_Dynamo_Equal(t *testing.T) {
	checkDynamo(t)

	table, err := ConnectDB()
	if err != nil {
		t.Fatalf(err.Error())
	}

	var samples []interface{}
	for i := 0; i < 10; i++ {
		samples = append(samples, Sample{
			PK:   fmt.Sprintf("%d", 1),
			SK:   fmt.Sprintf("SK_%d", i+1),
			Name: fmt.Sprintf("Name_%d", i+1),
		})
	}
	_, err = table.Batch().Write().Put(samples...).Run()
	if err != nil {
		t.Fatalf(err.Error())
	}

	b := NewBuilder()
	b.Equal("Name", "Name_1")

	var sample []Sample
	err = table.Get("PK", "1").
		Filter(b.JoinAnd(), b.Arg...).
		All(&sample)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(sample) != 1 ||
		sample[0].Name != "Name_1" {
		t.Fatalf("Not match sample: actual:%#v", sample)
	}
}
