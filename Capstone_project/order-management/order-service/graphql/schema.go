package graphql

import (
	"github.com/graphql-go/graphql"
	"order-management/order-service/model"
	"order-management/order-service/service"
)

func NewSchema(s service.OrderService) (graphql.Schema, error) {
	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id":            &graphql.Field{Type: graphql.Int},
			"customerId":    &graphql.Field{Type: graphql.Int},
			"productId":     &graphql.Field{Type: graphql.Int},
			"productName":   &graphql.Field{Type: graphql.String},
			"quantity":      &graphql.Field{Type: graphql.Int},
			"price":         &graphql.Field{Type: graphql.Float},
			"status":        &graphql.Field{Type: graphql.String},
			"reservationId": &graphql.Field{Type: graphql.String},
			"stockReleased": &graphql.Field{Type: graphql.Boolean},
			"createdAt":     &graphql.Field{Type: graphql.DateTime},
			"updatedAt":     &graphql.Field{Type: graphql.DateTime},
		},
	})

	root := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"order": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := int64(p.Args["id"].(int))
					return s.Get(p.Context, id)
				},
			},
			"orders": &graphql.Field{
				Type: graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return s.List(p.Context)
				},
			},
		},
	})

	input := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateOrderInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"customerId": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"productId":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"quantity":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	mut := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(input)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					x := p.Args["input"].(map[string]interface{})
					r := &model.CreateOrderRequest{
						CustomerID: int64(x["customerId"].(int)),
						ProductID:  int64(x["productId"].(int)),
						Quantity:   x["quantity"].(int),
					}
					return s.Create(p.Context, r)
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: root, Mutation: mut})
}
