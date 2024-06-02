package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gobase-lambda/db/mongo"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

var mongoURI = "mongodb://localhost:27017"

type TestVal struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	IntVal  int64              `bson:"intVal"`
	DeciVal decimal.Decimal    `bson:"deciVal"`
	StrVal  string             `bson:"strVal"`
	BoolVal bool               `bson:"boolVal"`
	TimeVal time.Time          `bson:"timeVal"`
}

func TestMongoCollection(t *testing.T) {
	coll, err := mongo.NewDefaultCollection(context.TODO(), mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	val1, _ := decimal.NewFromString("123.1232")
	val2, _ := decimal.NewFromString("123.1232")
	if err != nil {
		t.Fatal(err)
	}
	coll.InsertOne(map[string]interface{}{
		"strVal":  "value1",
		"intVal":  123,
		"deciVal": val1.Add(val2),
		"timeVal": time.Now().In(utils.IST),
	})
	cur := coll.FindOne(map[string]string{"strVal": "value1"})
	val := &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	coll.UpdateByID(val.ID, map[string]map[string]interface{}{"$set": {"strVal": "val2"}})
	cur = coll.FindOne(map[string]string{"strVal": "val2"})
	val = &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	coll.DeleteOne(map[string]string{"_id": val.ID.String()})
	cur = coll.FindOne(map[string]string{"_id": val.ID.String()})
	err = cur.Decode(val)
	if err != nil {
		fmt.Println(err)
	}

}

func TestMongoCollectionFindOne(t *testing.T) {
	coll, err := mongo.NewDefaultCollection(context.TODO(), mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	cur := coll.FindOne(map[string]string{"strVal": "value1"})
	val := &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
}

func TestMongoCollectionFindFetch(t *testing.T) {
	coll, err := mongo.NewDefaultCollection(context.TODO(), mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	loader := func(count int) []interface{} {
		val := make([]interface{}, count)
		for i := 0; i < count; i++ {
			val[i] = &TestVal{}
		}
		return val
	}
	data, err := coll.FindFetch(nil, loader)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range data {
		fmt.Printf("%+v\n", val)
	}
}

type Address struct {
	AddressLine1 string `bson:"addressLine1"`
	AddressLine2 string `bson:"addressLine2"`
	AddressLine3 string `bson:"addressLine3"`
	State        string `bson:"state"`
	PIN          string `bson:"pin"`
	Country      string `bson:"country"`
}

type PIITestVal struct {
	ID      primitive.ObjectID `bson:"_id"`
	DOB     string             `bson:"dob"`
	Name    string             `bson:"name"`
	Pan     string             `bson:"pan"`
	Email   string             `bson:"email"`
	Address Address            `bson:"address"`
}

func TestMongoCollectionPII(t *testing.T) {
	file, err := os.Open("piischeme.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	schemeByte, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	scheme := string(schemeByte)
	ctx := context.TODO()
	kmsArn := "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060"
	keyVaultNamespace := "encryption.__keyVault"
	keyAltName := "MongoPIITestKey"
	kmsProvider, err := mongo.GetDefaultAWSProvider(kmsArn)
	if err != nil {
		t.Fatal(err)
	}
	err = mongo.SetEncryptionKey(ctx, &scheme, mongoURI, keyVaultNamespace, keyAltName, log.GetDefaultLogger(), kmsProvider)
	if err != nil {
		t.Fatal(err)
	}
	coll, err := mongo.NewDefaultCollection(ctx, mongoURI, "GOLANGTEST", "PII")
	piicoll, err := mongo.NewDefaultCSFLECollection(ctx, mongoURI, kmsArn, keyVaultNamespace, "GOLANGTEST", "PII", scheme, []string{"pan", "email"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = piicoll.InsertOneWithHash(map[string]interface{}{
		"dob":   "1991-08-02",
		"name":  "Whats my name",
		"pan":   "ABCDE1234F",
		"email": "iam@go.com",
		"address": map[string]string{
			"addressLine1": "door no with street name",
			"addressLine2": "taluk and postal office",
			"addressLine3": "Optional landmark",
			"state":        "Tamil Nadu",
			"pin":          "636351",
			"country":      "India",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cur := piicoll.FindOneWithHash(map[string]interface{}{"pan": "ABCDE1234F"})
	val := &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	piicoll.UpdateByIDWithHash(val.ID, map[string]map[string]interface{}{"$set": {"email": "iam2@go.com"}})
	cur = piicoll.FindOneWithHash(map[string]interface{}{"email": "iam2@go.com"})
	val = &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	cur = piicoll.FindOne(map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}

	data, err := coll.Find(map[string]map[string]interface{}{"pan": {"$exists": true}})
	if err != nil {
		t.Fatal(err)
	}
	for data.Next(ctx) {
		decodeData := make(map[string]interface{})
		data.Decode(&decodeData)
		fmt.Printf("%+v\n", decodeData)
	}
	res, err := piicoll.DeleteOne(map[string]interface{}{"_id": val.ID})
	if err != nil {
		t.Fatal(err)
	}
	if res.DeletedCount != 1 {
		t.Fatal("Delete count is not matching")
	}
	cur = piicoll.FindOne(map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	if err != nil {
		fmt.Println(err)
	} else {
		t.Fatal(fmt.Errorf("doc shouldn't exist"))
	}

}
