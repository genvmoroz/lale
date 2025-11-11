// MongoDB initialization script for stress testing environment
// This script creates indexes for the cards collection

db = db.getSiblingDB(process.env.SERVICE_APP_MONGO_CARD_DATABASE);

print('Creating indexes for cards collection...');

// Create compound index for WordsExist and GetCardsByWords queries
// This index optimizes queries that filter by userid and search in wordinformationlist.word array
db.getCollection(process.env.SERVICE_APP_MONGO_CARD_COLLECTION).createIndex(
    {
        "userid": 1,
        "wordinformationlist.word": 1
    },
    {
        name: "userid_words_idx",
        background: true
    }
);

print('Index userid_words_idx created successfully');

// Create index for GetCardsForUser query
db.getCollection(process.env.SERVICE_APP_MONGO_CARD_COLLECTION).createIndex(
    {
        "userid": 1
    },
    {
        name: "userid_idx",
        background: true
    }
);

print('Index userid_idx created successfully');

// Create index for card ID lookups (used in SaveCards and DeleteCard)
db.getCollection(process.env.SERVICE_APP_MONGO_CARD_COLLECTION).createIndex(
    {
        "id": 1
    },
    {
        name: "id_idx",
        unique: true,
        background: true
    }
);

print('Index id_idx created successfully');

print('All indexes created successfully!');

