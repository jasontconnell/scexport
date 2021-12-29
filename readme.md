***scexport***

This was written to be used for conversion from Sitecore to another CMS, but file serialization, versus hitting the source database for an import, is much more universally accessible. So this can be used for any scenario. This is because the resulting files can be sent, rather than a database backup or having to open connections to the source database. Also any platform can read xml files and base64.

Usage: `scexport -c config.json -settings settings.json`

Yes there is a difference between configuration and settings :)  Configuration is more of a place for global settings, the settings is more local.

In this instance, the configuration will hold the connection string and a global template filter (to trim down on runtime), while the settings can be for a specific set of templates that you want grouped together in a way, e.g. Blog Posts or News Articles.

***Configuration JSON***

```
{
    "connectionString": "user id=user;password=pwd;server=dbserver;Database=database_Master",
    "globalTemplateFilter": [
        "/sitecore/templates/Company/",
        "/sitecore/templates/User Defined/Common/",
        "/sitecore/templates/System/Media/"
    ]
}
```

The connection string is the standard connection string. Global Template Filter will filter down the templates and fields that it has to load in order to trim down on the time it takes to run. It went from a minute to about 30 seconds when I implemented that part.

***Settings JSON***

```
{
    "filterLanguage": "en",
    "templates": [
        {
            "name": "blog",
            "templateId": "AAAAAAAA-BBBB-CCCC-DDDD-123456789ABC",
            "path": "/sitecore/content/home/blog",
            "fields": [
                {
                    "name": "PostTitle"
                },
                {
                    "name": "Author"
                },
                {
                    "name": "PostDate"
                },
                {
                    "name": "BodyText"
                },
                {
                    "name": "TeaserText"
                },
                {
                    "name": "Category",
                    "refField": "CategoryName"
                },
                {
                    "name": "Image"
                },
                {
                    "name": "Video",
                    "alias": "VideoURL",
                    "refField": "VideoUrl"
                },
                {
                    "name": "Video",
                    "alias": "VideoBlurb",
                    "refField": "VideoBlurb"
                }
            ]
        }
    ],
    "output": {
        "contentFormat": "xml",
        "contentLocation": "./output/blog/",
        "blobLocation": "./output/blog/blobs/"
    }
}
```

This includes the language filter, the template, and the fields that you want to export. Only xml is supported at the moment. `scexport` will write out the data to the locations specified in the `output` section.

No settings are required right now for standard fields, but a `Droplink` for instance, will require a field name that should be output.

If a field references an object and you want to use more than one field from the referenced data, use the "alias" to specify how it will be output. Alias is only used for output.

***Output***

`scexport` will output one file for the contents. So in this example, all of the blog posts will be in a `blog.xml` file in the specified output folder. Example xml is below. I've only included the interesting bits (rich text and blobs).

```
 <items>
  <item type="blog" name="test-blog-post-1">
   <fields>
    <field name="BodyText"><!CDATA[[
        <p>body text</p>
        <p>here is an image referenced within sitecore:</p>
        <img src="blobref:abcdabcd-abcd-defa-1234-123456789123" alt="Awesome image of the meaning of life" width="1200" height="700" />
    ]]></field>
   </fields>
   <blobrefs>
    <blob id="abcdabcd-abcd-defa-1234-123456789123" filename="The_meaning_of_life-1200x700.jpg"></blob>
   </blobrefs>
  </item>
</items>
```

As you can see, the images that link to an image within Sitecore (like `-/media/ABCDABCDABCDDEFA1234123456789123.ashx`) will be pulled and placed into the blobs folder, and referenced here. Image fields will be handled similarly.

All blobs will be output to the specified folder in the output section. They will be one file per blob, different from how content is handled. The blob xml will look like this:

```
<blob id="abcdabcd-abcd-defa-1234-123456789123" filename="The_meaning_of_life-1200x700.jpg" length="42424242">
    ... base64 encoded binary data ...
</blob>
```

The blob xml filename will be the blob filename with .xml appended, in this example, `The_meaning_of_life-1200x700.jpg.xml`
