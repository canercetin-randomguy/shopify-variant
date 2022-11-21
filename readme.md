Tool to add variants to products on Shopify.
Main loop: 
- Iterate over products.json that is retrieved from RetrieveProducts, twice.
- If first iteration product and second iteration products type is the same, add second iteration product
as variant to the first iteration product.
- This is done by hitting admin/api/2022-10/products/%d/images.json and uploading image to desired product.
- Then hitting admin/api/2022-10/variants/%d.json to add variant, image src is retrieved from previous steps response.