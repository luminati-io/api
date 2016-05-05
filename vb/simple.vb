Imports System.Net

Class Client
    Inherits WebClient
    Const username As String = "lum-customer-CUSTOMER-zone-YOURZONE"
    Const password As String = "YOURPASS"
    Const port = 22225
    Public session_id As String = New Random().Next().ToString()

    Public Sub New(Optional country As String = Nothing)
        Me.Proxy = New WebProxy("zproxy.luminati.io", port)
        Dim login = username &
            If(country IsNot Nothing, "-country-" & country, "") &
            "-session-" & session_id
        Me.Proxy.Credentials = New NetworkCredential(login, password)
    End Sub

    Protected Overrides Function GetWebRequest(address As Uri) As WebRequest
        Dim request = MyBase.GetWebRequest(address)
        request.ConnectionGroupName = session_id
        Return request
    End Function
End Class

Module Example
    Sub Main()
        Console.WriteLine("Performing request(s)")
        Dim session as New Client()
        ' Put full scraping sequence below:
        Console.WriteLine(session.DownloadString("http://lumtest.com/myip.json"))
        ' client.DownloadString(...second request...);
    End Sub
End Module

