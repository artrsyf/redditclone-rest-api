var username = process.env.MONGODB_USER;
var password = process.env.MONGODB_PASSWORD;
var database = process.env.MONGODB_DATABASE;

print("Adding New Users");
db = db.getSiblingDB("admin");
db.createUser({
  user: username,
  pwd: password,
  roles: [{ role: "readWrite", db: database }],
});
print("End Adding the User Roles.");