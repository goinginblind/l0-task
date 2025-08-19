# Database Schema

The **Schema:** 

![Database Schema](schema.png)
1. **Why are there four tables?**
I split the data into four tables to keep some level of normalization. While the example JSON had overlapping fields (e.g. `transaction` in `payments` and `order_uid` in `orders`), I didn’t assume they’re always the same. So I treated them as independent fields instead of forcing them into one.

2. **Why the use of surrogate keys?**
I tought that maybe, if the db stores i.e. 10 million records, the joins using strings (`TEXT`) as primary keys will lead to some perfomance issues. Using surrogate integer IDs reduces storage size and makes joins faster, since integers are more efficient for the CPU to compare.
_Alternative_: I could use natural keys (order_uid, transaction) directly as primary keys. We loose an extra column, but get slower joins and more storage overhead in the long run.

3. **Why NOT NULL constraints everywhere?**
This ensures the database rejects bad/incomplete data if the Go application logic slips. It is a safety net

### Other Documentation:
* [Consumer Decision Tree](consumer.md)
* [Cache Implementation](cache.md)
* [JSON Validation](validation.md)
* [Errors, Metrics, and DB Health Checks](misc.md)

### Back to [Main README](../README.md)
