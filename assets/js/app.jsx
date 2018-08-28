class Synchronizer extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            nokia: "Loading",
            fatsecret: "Loading"
        };
    }

    componentDidMount() {
        this.syncStatus();
    }

    syncNokia() {
        var url = window.location.href + "syncNokia";
        fetch(url, {
            "cache": "no-cache",
        }).then(
            response => response.json()
        ).then(
            json => window.location = json.url
        );
    }

    render() {

        return (
            <table border="1" width="50%" cellPadding="5">
                <tbody>
                    <tr>
                        <td>Sync state of Nokia</td>
                        <td>{this.state.nokia}</td>
                        <td><input type="button" value="Link"
                                   onClick={this.syncNokia}/></td>
                    </tr>
                    <tr>
                        <td>Sync state of FatSecret</td>
                        <td>{this.state.fatsecret}</td>
                        <td><input type="button" value="Link"/></td>
                    </tr>
                </tbody>
            </table>
        )
    }

    syncStatus() {
        // call server to update sync progress and check statuses
        var url = window.location.href + "syncStatus";
        fetch(url, {
            "cache": "no-cache"
        }).then(
            response => response.json()
        ).then(
            json => this.updateStatus(json)
        );
    }

    updateStatus(status) {

        function value(status) {
            if (status) {
                return "Linked";
            } else {
                return "Not Linked";
            }
        }

        this.setState({
            nokia: value(status.nokia),
            fatsecret: value(status.fatsecret)
        });
    }

  
}

ReactDOM.render(
      <Synchronizer/>,
      document.getElementById('root')
);
