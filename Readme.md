# Vendor Mailer
This program downloads the current inventory and emails individual vendors with their inventory levels.
It utilizes https://mandrill.com to send emails and therefore requires signing up for this service. 
Mandrill offers 12,000 free emails monthly which should be more than enough for this use case.

The latest version of this Readme and the program itself can be found at https://github.com/JFMarket/vendor-mailer.

## Installation
1. Signup for Mandrill
1. Install program
1. Setup a scheduled task

### Signup for Mandrill
1. Navigate to https://mandrill.com/signup/ in your web browser.
1. Enter your email address.
1. Choose a password for the Mandrill service. (This form is not asking for your actual email password.)
1. After signing up, if not logged in automatically, navigate to https://mandrillapp.com/ and sign in.
1. Navigate to https://mandrillapp.com/settings/index
1. Click the "New API Key" button toward the bottom of the page.
1. Give it a short description if requested and click "Create API Key" (The checkboxes should all be blank)
1. The key will appear at the bottom of the page. Write it down you will need it soon.
1. Congratulations, you have just signed up for Mandrill. No more changes should be necessary for Mandrill settings.

### Install Program
If you have a Go environment setup, simply `go get github.com/JFMarket/vendor-mailer`.
If you don't know what that is, keep reading.

#### On Windows
You should have received a zip file with a name similar to "vendor-mailer.zip".

1. Extract this zip file somewhere that won't accidentally be deleted. (My Documents should be fine. Your Desktop might not be a good idea.)
1. If you have extracted this file to "My Documents" then you should now have a "vendor-mailer" directory inside "My Documents" that contains
a "dist" directory and a file named "vendor-emails-example.csv".
1. Rename "vendor-emails-example.csv" to "vendor-emails.csv".
  * "vendor-emails.csv" maps vendor names to their email address.
1. Open "vendor-emails.csv" with Excel and insert vendor names as they appear in ShopKeep.
They must match exactly or they will not receive an email. In the email column, as you would guess,
enter the email address of the corresponding vendor.
  * You will have to update this file as new vendors join.
  * You do not have to do anything for changes to this file to be recognized.
  The next time the program runs, it will use the latest version.
1. Vendor Mailer is now configured.

### Setup a Scheduled Task
#### On Windows
Windows has a built-in task scheduler that will run this program periodically. Every time this program runs,
it will download the inventory and send out emails to vendors. While possible to run as often as you like, 
it is recommended that you do not run it more frequently than once a week. Vendors may not like receiving
emails too often. This task looks complicated, but you should only have to do it once and then you won't have
to worry about it.

1. Open Task Scheduler by clicking the Start button, clicking Control Panel, clicking System and Security, clicking Administrative Tools, and then double-clicking Task Scheduler.‌
If you're prompted for an administrator password or confirmation, type the password or provide confirmation.
1. On the "Actions" panel, click "Create Basic Task".
1. Name the task "Vendor Mailer". A description is optional. Click "Next".
1. Select "Weekly". Click "Next".
1. Set the start date and the time you want the task to run. Remember, your computer must be turned on for the task to run.
1. The default of 1 for "Recur every" means that the task will run every week. Setting this value to 2 would mean every two
weeks and so on. Change this if necessary.
1. Select the day of the week to run the task on. Selecting more than one day will result in the task running multiple times
a week. This is not recommended. Click "Next".
1. Select "Start a program". Click "Next".
1. Click "Browse". Navigate to the directory where you extracted the zip file. Double click "dist". Double click `vendor-mailer_windows_amd64.exe`.
1. In the "Add arguments" text field, enter the following: `-email=myemail@gmail.com -password=whatevermypasswordis -vendorEmails=..\vendor-emails.csv -key=mymandrillapikey -fromEmail=theemailisignedupwithonmandrill@mail.com -fromName=Sender Name`
  * Parameters Explanation:
    * -email is the email address used to login to ShopKeep.
    * -password is the password used to login to ShopKeep.
    * -vendorEmails is the CSV file configured above that maps Vendors to their emails.
    * -key is the API key created above for Mandrill.
    * -fromEmail is the email address used to signup for Mandrill and will be displayed in the From field of each email.
    * -fromName is the name that recipients will see associated with the sending address.
1. In the "Start in" text field, copy and paste the path in "Program/Script" from C:\ to \dist. Do not include the quotation mark before the C: or the `\` after dist.
  * If "Program/script" contained `"C:\Users\Bob\Documents\vendor-mailer\dist\vendor-mailer_windows_amd64.exe"`, then "Start in" should contain `C:\Users\Bob\Documents\vendor-mailer\dist`.
1. Check the box for "Open the Properties dialog for this task..."
1. Click "Finish".
1. In the Properties dialog, 
  * Under "Security Options", Select "Run whether the user is logged on or not"
  * On the "Conditions" tab under "Power"
    1. Uncheck "Start the task only if the computer is on AC power"
    1. Check "Wake the computer to run this task."
  * Click "OK"
1. Enter your password so the task can run if you are not logged in.
1. You can now close the Task Scheduler window.
1. Vendor emails will now be sent out at the scheduled time and day of each week.

#### On Linux
1. Use cron. The command line above applies on Linux. Just switch `\` and `/` and quote flags appropriately.