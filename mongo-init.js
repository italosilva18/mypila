// MongoDB initialization script
// This script runs when the MongoDB container starts for the first time

db = db.getSiblingDB('m2m_financeiro');

// Create the transactions collection with a validator
db.createCollection('transactions', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['month', 'year', 'amount', 'category', 'status'],
      properties: {
        month: {
          bsonType: 'string',
          description: 'Month name or reference (required)'
        },
        year: {
          bsonType: 'int',
          description: 'Year (required)'
        },
        amount: {
          bsonType: 'double',
          description: 'Transaction amount (required)'
        },
        category: {
          bsonType: 'string',
          enum: ['Salário', 'Férias', 'Custos de IA', 'Custo de Docker'],
          description: 'Category (required)'
        },
        status: {
          bsonType: 'string',
          enum: ['PAGO', 'ABERTO'],
          description: 'Status (required)'
        },
        description: {
          bsonType: 'string',
          description: 'Optional description'
        }
      }
    }
  }
});

// Create indexes for better query performance
db.transactions.createIndex({ month: 1, year: 1 });
db.transactions.createIndex({ category: 1 });
db.transactions.createIndex({ status: 1 });

print('MongoDB initialization completed!');
