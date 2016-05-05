Imports System.Net
Imports System.Threading.Tasks

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
        ' Example of how to run 100 parallel sessions (100 different Exit Node
        ' IPs) by using 100 threads, each with a unique session_id.
        Parallel.For(0, 100,
        Sub(i)
            Dim session as New Client()
            Console.WriteLine("Performing request(s) from session #{0}", i)
            ' Put complete scraping sequence that should be done in one session (from one IP):
            Console.WriteLine(session.DownloadString("http://lumtest.com/myip.json"))
            ' session.DownloadString(...second request...);
        End Sub)
    End Sub
End Module

