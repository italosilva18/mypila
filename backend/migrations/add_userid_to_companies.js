// MongoDB Migration Script
// Adds userId field to existing companies
//
// Usage:
//   mongosh mongodb://localhost:27017/m2m --file add_userid_to_companies.js
//
// OR from mongosh:
//   use m2m
//   load("add_userid_to_companies.js")

// ============================================
// CONFIGURATION
// ============================================

// Set this to a valid user ID from your users collection
// You can get one with: db.users.findOne()._id
const DEFAULT_USER_ID = ObjectId("000000000000000000000000"); // CHANGE THIS!

// ============================================
// MIGRATION SCRIPT
// ============================================

print("=== Company UserID Migration Script ===");
print("");

// Check if users collection exists
const usersCount = db.users.countDocuments();
print(`Found ${usersCount} users in database`);

if (usersCount === 0) {
    print("WARNING: No users found in database!");
    print("Please create at least one user before running this migration.");
    print("Aborting migration.");
    quit(1);
}

// Get the first user as default
const firstUser = db.users.findOne();
const defaultUserId = firstUser._id;
print(`Using default userId: ${defaultUserId}`);
print("");

// Find companies without userId
const companiesWithoutUserId = db.companies.countDocuments({ userId: { $exists: false } });
print(`Found ${companiesWithoutUserId} companies without userId field`);

if (companiesWithoutUserId === 0) {
    print("No companies need migration. All companies already have userId.");
    quit(0);
}

// Ask for confirmation (comment out if running non-interactively)
print("");
print("This will add userId to all companies without it.");
print("Press Ctrl+C to cancel or Enter to continue...");
// For non-interactive mode, comment out the line below:
// prompt();

// Perform the migration
print("");
print("Starting migration...");

const result = db.companies.updateMany(
    { userId: { $exists: false } },
    { $set: { userId: defaultUserId } }
);

print(`Updated ${result.modifiedCount} companies`);

// Verify the migration
const remainingWithoutUserId = db.companies.countDocuments({ userId: { $exists: false } });
print(`Remaining companies without userId: ${remainingWithoutUserId}`);

if (remainingWithoutUserId === 0) {
    print("");
    print("✓ Migration completed successfully!");
} else {
    print("");
    print("⚠ Warning: Some companies still don't have userId");
}

// Show statistics
print("");
print("=== Migration Statistics ===");
const companiesByUser = db.companies.aggregate([
    { $group: { _id: "$userId", count: { $sum: 1 } } },
    { $lookup: { from: "users", localField: "_id", foreignField: "_id", as: "user" } },
    { $unwind: { path: "$user", preserveNullAndEmptyArrays: true } }
]);

print("");
print("Companies per user:");
companiesByUser.forEach(stat => {
    const userName = stat.user ? stat.user.name : "Unknown User";
    const userEmail = stat.user ? stat.user.email : "N/A";
    print(`  ${userName} (${userEmail}): ${stat.count} companies`);
});

print("");
print("=== Next Steps ===");
print("1. Create index for performance:");
print("   db.companies.createIndex({ userId: 1 })");
print("");
print("2. If you need to assign companies to specific users:");
print("   db.companies.updateMany(");
print("     { name: 'Company Name' },");
print("     { $set: { userId: ObjectId('user_id_here') } }");
print("   )");
print("");
