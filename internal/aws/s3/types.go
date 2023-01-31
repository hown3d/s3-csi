package s3

import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "time"
)

type ObjectAttrs struct {
    Size       uint64
    ModifyTime *time.Time
    AccessTime *time.Time
    ChangeTime *time.Time
}

func (a *ObjectAttrs) toTagSlice() []types.Tag {
    var tags []types.Tag
    if a.ChangeTime != nil {
        tags = append(tags, types.Tag{
            Key:   aws.String(CHANGE_TIME_METADATA_KEY),
            Value: aws.String(a.ChangeTime.Format(time.UnixDate)),
        })
    }
    return tags
}
