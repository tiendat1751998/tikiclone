# Business Rules & Voucher Calculation Engines

To match Tiki's checkout experience, the platform enforces strict business rules regarding voucher stacks.

## 1. The Stackable Voucher Calculation Formula
When checking out, buyers can stack vouchers in this exact sequence:

$$	ext{Final Price} = ((	ext{Base Item Price} - 	ext{Shop Voucher}) - 	ext{Platform Voucher}) - 	ext{Shipping Discount} - 	ext{Tiki Coins}$$

### Stackable Constraint Matrix
- **Shop Voucher**: Provided by sellers. Applies only to specific shop products.
- **Platform Voucher**: Provided by Tiki. Can be fixed-amount or percentage discount (e.g. 10% off total bill).
- **Shipping Discount**: Subsidizes shipping costs.
- **Tiki Coins**: Maximum 50% value of the remaining order.

## 2. Order Cancellation & Stock Restoral Timelines
- Orders in `PENDING_PAYMENT` state have a maximum lifecycle of **15 minutes** (5 minutes during campaigns).
- If unpaid within this window, a cron job executes Order cancel, triggering the compensatory Kafka event `order.cancelled`.
- Inventory consumer catches this event and releases the reserved stock back into Redis and PostgreSQL.
