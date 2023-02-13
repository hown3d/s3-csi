package s3

import (
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "strconv"
)

type Metadata map[string]string

func defaultObjectMetadata() Metadata {
    return map[string]string{
        CHANGE_TIME_NSEC_METADATA_KEY: strconv.Itoa(int(TimeNowFunc().Nanosecond())),
        CHANGE_TIME_SEC_METADATA_KEY:  strconv.Itoa(int(TimeNowFunc().Unix())),
    }
}

func mergeMetadata(m1, m2 Metadata) Metadata {
    if m2 == nil {
        return m1
    }
    for key, val := range m1 {
        m2[key] = val
    }
    return m2
}

func metadataFromTagSlice(tags []types.Tag) Metadata {
    var metadata Metadata
    for _, tag := range tags {
        metadata[*tag.Key] = *tag.Value
    }
    return metadata
}

func (m Metadata) toTagSlice() []types.Tag {
    tagSlice := make([]types.Tag, 0, len(m))
    for key, val := range m {
        tagSlice = append(tagSlice, types.Tag{
            Key:   &key,
            Value: &val,
        })
    }
    return tagSlice
}
