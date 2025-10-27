package s3_bucket

type S3BucketCommands struct {
	Add    AddCmd    `cmd:"" help:"Add an S3 bucket configuration"`
	Delete DeleteCmd `cmd:"" help:"Delete an S3 bucket configuration" aliases:"rm,delete"`
	List   ListCmd   `cmd:"" help:"List S3 bucket configurations" aliases:"ls"`
	Update UpdateCmd `cmd:"" help:"Update an S3 bucket configuration"`
}
